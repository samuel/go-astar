package main

import (
	"errors"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/samuel/go-astar"
)

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

type ImageMap struct {
	Pix              []byte
	YStride, XStride int
	Width, Height    int
	Stddev           float64

	setter func(x, y int, c color.Color)
}

func colorDiff(c1, c2 byte) float64 {
	a := abs(int(c1) - int(c2))
	return float64(a * a)
}

func NewImageMap(img image.Image) (*ImageMap, error) {
	var im *ImageMap
	switch m := img.(type) {
	case *image.YCbCr:
		im = &ImageMap{m.Y, m.YStride, 1, img.Bounds().Dx(), img.Bounds().Dy(), 1.0, nil}
		var verticalRes, horizontalRes int
		switch m.SubsampleRatio {
		case image.YCbCrSubsampleRatio420:
			verticalRes = 2
			horizontalRes = 2
		case image.YCbCrSubsampleRatio422:
			verticalRes = 1
			horizontalRes = 2
		case image.YCbCrSubsampleRatio440:
			verticalRes = 2
			horizontalRes = 1
		case image.YCbCrSubsampleRatio444:
			verticalRes = 1
			horizontalRes = 1
		default:
			return nil, errors.New("unsupported YCbCr subsample ratio")
		}
		im.setter = func(x, y int, c color.Color) {
			r, g, b, _ := c.RGBA()
			yc, cb, cr := color.RGBToYCbCr(uint8(r>>8), uint8(g>>8), uint8(b>>8))
			m.Y[y*m.YStride+x] = yc
			off := y/verticalRes*m.CStride + x/horizontalRes
			m.Cb[off] = cb
			m.Cr[off] = cr
		}
	case *image.RGBA:
		im = &ImageMap{m.Pix[1:], m.Stride, 4, img.Bounds().Dx(), img.Bounds().Dy(), 1.0, m.Set}
	case *image.Gray:
		im = &ImageMap{m.Pix, m.Stride, 1, img.Bounds().Dx(), img.Bounds().Dy(), 1.0, m.Set}
	default:
		return nil, errors.New("Unsupported image format")
	}

	m := -1.0
	s := 0.0
	count := 0
	for y := 0; y < im.Height; y++ {
		for x := 0; x < im.Width; x++ {
			count++
			v := float64(im.Pix[y*im.YStride+x*im.XStride])
			oldM := m
			if oldM == -1 {
				m = v
				s = 0
			} else {
				m = oldM + ((v - oldM) / float64(count))
				s += (v - oldM) * (v - m)
			}
		}
	}
	stddev := math.Sqrt(s / float64(count-1))
	im.Stddev = stddev

	return im, nil
}

func (im *ImageMap) Neighbors(node int) ([]astar.Edge, error) {
	edges := make([]astar.Edge, 0, 8)

	x := node % im.Width
	y := node / im.Width
	off := y*im.YStride + x*im.XStride
	c := im.Pix[off]

	if x > 0 {
		edges = append(edges, astar.Edge{node - 1, colorDiff(c, im.Pix[off-im.XStride])})
		if y > 0 {
			edges = append(edges, astar.Edge{node - 1 - im.Width, colorDiff(c, im.Pix[off-im.XStride-im.YStride])})
		}
		if y < im.Height-1 {
			edges = append(edges, astar.Edge{node - 1 + im.Width, colorDiff(c, im.Pix[off-im.XStride+im.YStride])})
		}
	}
	if x < im.Width-1 {
		edges = append(edges, astar.Edge{node + 1, colorDiff(c, im.Pix[off+im.XStride])})
		if y > 0 {
			edges = append(edges, astar.Edge{node + 1 - im.Width, colorDiff(c, im.Pix[off+im.XStride-im.YStride])})
		}
		if y < im.Height-1 {
			edges = append(edges, astar.Edge{node + 1 + im.Width, colorDiff(c, im.Pix[off+im.XStride+im.YStride])})
		}
	}
	if y > 0 {
		edges = append(edges, astar.Edge{node - im.Width, colorDiff(c, im.Pix[off-im.YStride])})
	}
	if y < im.Height-1 {
		edges = append(edges, astar.Edge{node + im.Width, colorDiff(c, im.Pix[off+im.YStride])})
	}
	return edges, nil
}

func (im *ImageMap) HeuristicCost(start int, end int) (float64, error) {
	endY := end / im.Width
	endX := end % im.Width
	startY := start / im.Width
	startX := start % im.Width
	a := abs(endY - startY)
	b := abs(endX - startX)
	return math.Sqrt(float64(a*a+b*b)) * im.Stddev, nil
}

func (im *ImageMap) Set(x, y int, c color.Color) {
	im.setter(x, y, c)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("syntax: imagepath [path]")
	}
	rd, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer rd.Close()
	img, _, err := image.Decode(rd)
	if err != nil {
		log.Fatal(err)
	}

	im, err := NewImageMap(img)
	if err != nil {
		log.Fatal(err)
	}
	path, err := astar.FindPath(im, 0, img.Bounds().Dx()-1+img.Bounds().Dx()*(img.Bounds().Dy()-1))
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range path {
		x := node % img.Bounds().Dx()
		y := node / img.Bounds().Dx()
		im.Set(x, y, color.RGBA{255, 0, 0, 255})
	}
	wr, err := os.Create("out.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(wr, img); err != nil {
		log.Fatal(err)
	}
	wr.Close()
}
