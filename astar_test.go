package astar

import (
	"fmt"
	"math"
	"testing"
)

const (
	sqrt2 = 1.4142135623730951
)

type gridMap struct {
	grid   []int
	width  int
	height int
}

func abs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}

func (g *gridMap) Neighbors(node int) []Edge {
	edges := make([]Edge, 0, 8)
	addNode := func(x, y int, cost float64) {
		v := g.grid[y*g.width+x]
		if v == 0 {
			edges = append(edges, Edge{y*g.width + x, cost})
		}
	}

	y := node / g.width
	x := node % g.width

	if x > 0 {
		addNode(x-1, y, 1)
		if y > 0 {
			addNode(x-1, y-1, sqrt2)
		}
		if y < (g.height - 1) {
			addNode(x-1, y+1, sqrt2)
		}
	}
	if x < (g.width - 1) {
		addNode(x+1, y, 1)
		if y > 0 {
			addNode(x+1, y-1, sqrt2)
		}
		if y < (g.height - 1) {
			addNode(x+1, y+1, sqrt2)
		}
	}
	if y > 0 {
		addNode(x, y-1, 1)
	}
	if y < (g.height - 1) {
		addNode(x, y+1, 1)
	}
	return edges
}

func (g *gridMap) HeuristicCost(start int, end int) float64 {
	endY := end / g.width
	endX := end % g.width
	startY := start / g.width
	startX := start % g.width
	a := abs(endY - startY)
	b := abs(endX - startX)
	return math.Sqrt(float64(a*a + b*b))
}

func TestAstar(t *testing.T) {
	mp := &gridMap{
		grid: []int{
			0, 0, 0, 0, 1, 0, 0, 0, 0, 0,
			0, 1, 1, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 1, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 1, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 1, 0, 1, 0, 0, 1, 1, 0,
			0, 0, 1, 0, 1, 0, 0, 0, 1, 0,
			0, 0, 1, 0, 1, 0, 0, 1, 0, 0,
			1, 1, 1, 0, 1, 0, 1, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		},
		width:  10,
		height: 10,
	}
	path := FindPath(mp, 5*mp.width, 3*mp.width+9)
	fmt.Printf("%+v\n", path)
	for y := 0; y < mp.height; y++ {
		for x := 0; x < mp.width; x++ {
			o := y*mp.width + x
			pth := false
			for _, p := range path {
				if p == o {
					fmt.Printf(".")
					pth = true
					break
				}
			}
			if !pth {
				if mp.grid[y*mp.width+x] == 0 {
					fmt.Printf(" ")
				} else {
					fmt.Printf("#")
				}
			}
		}
		fmt.Println()
	}
}

func BenchmarkFindPath(b *testing.B) {
	mp := &gridMap{
		grid: []int{
			0, 0, 0, 0, 1, 0, 0, 0, 0, 0,
			0, 1, 1, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 1, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 1, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 1, 0, 1, 0, 0, 1, 1, 0,
			0, 0, 1, 0, 1, 0, 0, 0, 1, 0,
			0, 0, 1, 0, 1, 0, 0, 1, 0, 0,
			1, 1, 1, 0, 1, 0, 1, 0, 0, 0,
			0, 0, 0, 0, 1, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		},
		width:  10,
		height: 10,
	}
	for i := 0; i < b.N; i++ {
		FindPath(mp, 5*mp.width, 3*mp.width+9)
	}
}
