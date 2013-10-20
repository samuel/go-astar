package astar

import (
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

func (g *gridMap) Neighbors(node int, edges []Edge) ([]Edge, error) {
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
	return edges, nil
}

func (g *gridMap) HeuristicCost(start int, end int) (float64, error) {
	endY := end / g.width
	endX := end % g.width
	startY := start / g.width
	startX := start % g.width
	a := abs(endY - startY)
	b := abs(endX - startX)
	return math.Sqrt(float64(a*a + b*b)), nil
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
	path, err := FindPath(mp, 5*mp.width, 3*mp.width+9)
	if err != nil {
		t.Fatal(err)
	}
	expected := []int{50, 40, 30, 20, 10, 1, 2, 13, 23, 33, 43, 53, 63, 73, 83, 94, 85, 86, 77, 68, 59, 49, 39}
	if len(path) < len(expected) {
		t.Fatalf("Expected a path length of %d instead of %d", len(expected), len(path))
	}
	for i, e := range expected {
		if path[i] != e {
			t.Fatalf("Expected node at path index %d to be %d instead of %d", i, e, path[i])
		}
	}
	for y := 0; y < mp.height; y++ {
		out := make([]byte, mp.width)
		for x := 0; x < mp.width; x++ {
			o := y*mp.width + x
			pth := false
			for _, p := range path {
				if p == o {
					out[x] = '.'
					pth = true
					break
				}
			}
			if !pth {
				if mp.grid[y*mp.width+x] == 0 {
					out[x] = ' '
				} else {
					out[x] = '#'
				}
			}
		}
		t.Logf(string(out))
	}
}

func TestImpossible(t *testing.T) {
	mp := &gridMap{
		grid: []int{
			0, 0, 0, 0, 1, 0, 0, 0, 0, 0,
			1, 1, 1, 0, 1, 0, 0, 0, 0, 0,
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
	_, err := FindPath(mp, 5*mp.width, 3*mp.width+9)
	if err != ErrImpossible {
		t.Fatal("Expected ErrImpossible when no path is possible")
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
