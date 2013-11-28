package main

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/samuel/go-astar/astar"
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

func colorCost(c1, c2 byte) float64 {
	a := abs(int(c1) - int(c2))
	return float64(a*a + 1)
}

func NewImageMap(img image.Image) (*ImageMap, error) {
	var im *ImageMap
	switch m := img.(type) {
	case *image.YCbCr:
		im = &ImageMap{
			Pix:     m.Y,
			YStride: m.YStride,
			XStride: 1,
			Width:   img.Bounds().Dx(),
			Height:  img.Bounds().Dy(),
			Stddev:  1.0,
		}
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
		im = &ImageMap{
			Pix:     m.Pix[1:],
			YStride: m.Stride,
			XStride: 4,
			Width:   img.Bounds().Dx(),
			Height:  img.Bounds().Dy(),
			Stddev:  1.0,
			setter:  m.Set,
		}
	case *image.Gray:
		im = &ImageMap{
			Pix:     m.Pix,
			YStride: m.Stride,
			XStride: 1,
			Width:   img.Bounds().Dx(),
			Height:  img.Bounds().Dy(),
			Stddev:  1.0,
			setter:  m.Set,
		}
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

func (im *ImageMap) Neighbors(node astar.Node, edges []astar.Edge) ([]astar.Edge, error) {
	x := int(node) % im.Width
	y := int(node) / im.Width
	off := y*im.YStride + x*im.XStride
	c := im.Pix[off]

	if x > 0 {
		edges = append(edges, astar.Edge{Node: node - 1, Cost: colorCost(c, im.Pix[off-im.XStride])})
		if y > 0 {
			edges = append(edges, astar.Edge{Node: node - 1 - astar.Node(im.Width), Cost: math.Sqrt2 * colorCost(c, im.Pix[off-im.XStride-im.YStride])})
		}
		if y < im.Height-1 {
			edges = append(edges, astar.Edge{Node: node - 1 + astar.Node(im.Width), Cost: math.Sqrt2 * colorCost(c, im.Pix[off-im.XStride+im.YStride])})
		}
	}
	if x < im.Width-1 {
		edges = append(edges, astar.Edge{Node: node + 1, Cost: colorCost(c, im.Pix[off+im.XStride])})
		if y > 0 {
			edges = append(edges, astar.Edge{Node: node + 1 - astar.Node(im.Width), Cost: math.Sqrt2 * colorCost(c, im.Pix[off+im.XStride-im.YStride])})
		}
		if y < im.Height-1 {
			edges = append(edges, astar.Edge{Node: node + 1 + astar.Node(im.Width), Cost: math.Sqrt2 * colorCost(c, im.Pix[off+im.XStride+im.YStride])})
		}
	}
	if y > 0 {
		edges = append(edges, astar.Edge{Node: node - astar.Node(im.Width), Cost: colorCost(c, im.Pix[off-im.YStride])})
	}
	if y < im.Height-1 {
		edges = append(edges, astar.Edge{Node: node + astar.Node(im.Width), Cost: colorCost(c, im.Pix[off+im.YStride])})
	}
	return edges, nil
}

func (im *ImageMap) HeuristicCost(start, end astar.Node) (float64, error) {
	endY := int(end) / im.Width
	endX := int(end) % im.Width
	startY := int(start) / im.Width
	startX := int(start) % im.Width
	a := abs(endY - startY)
	b := abs(endX - startX)
	return math.Sqrt(float64(a*a + b*b)), nil // * im.Stddev / 2, nil
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

	log.Println("Processing image")
	im, err := NewImageMap(img)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 2 {
		wr, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		defer wr.Close()
		if err := pprof.StartCPUProfile(wr); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Finding path")
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	totalAlloc := memStats.TotalAlloc
	t := time.Now()
	path, err := astar.FindPath(im, 0, astar.Node(img.Bounds().Dx()-1+img.Bounds().Dx()*(img.Bounds().Dy()-1)))
	pprof.StopCPUProfile()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("\t%d ms", time.Since(t).Nanoseconds()/1e6)
	runtime.ReadMemStats(&memStats)
	log.Printf("\t%d MB allocated", (memStats.TotalAlloc-totalAlloc)/(1024*1024))

	log.Printf("Nodes in path: %d", len(path))

	log.Println("Rendering path")
	for _, node := range path {
		x := int(node) % img.Bounds().Dx()
		y := int(node) / img.Bounds().Dx()
		im.Set(x, y, color.RGBA{0, 255, 0, 255})
	}

	log.Println("Encoding/writing output image")
	wr, err := os.Create("out.jpg")
	if err != nil {
		log.Fatal(err)
	}
	if err := jpeg.Encode(wr, img, nil); err != nil {
		log.Fatal(err)
	}
	wr.Close()
}
