package astar

import (
	"fmt"
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

func (g *gridMap) Neighbors(node Node) []Edge {
	edges := make([]Edge, 0, 8)
	addNode := func(x, y int, cost float64) {
		v := g.grid[y*g.width+x]
		if v == 0 {
			edges = append(edges, Edge{Node{x, y}, cost})
		}
	}

	if node.X > 0 {
		addNode(node.X-1, node.Y, 1)
		if node.Y > 0 {
			addNode(node.X-1, node.Y-1, sqrt2)
		}
		if node.Y < (g.height - 1) {
			addNode(node.X-1, node.Y+1, sqrt2)
		}
	}
	if node.X < (g.width - 1) {
		addNode(node.X+1, node.Y, 1)
		if node.Y > 0 {
			addNode(node.X+1, node.Y-1, sqrt2)
		}
		if node.Y < (g.height - 1) {
			addNode(node.X+1, node.Y+1, sqrt2)
		}
	}
	if node.Y > 0 {
		addNode(node.X, node.Y-1, 1)
	}
	if node.Y < (g.height - 1) {
		addNode(node.X, node.Y+1, 1)
	}
	return edges
}

func (g *gridMap) HeuristicCost(start Node, end Node) float64 {
	return float64(abs(end.Y-start.Y) + abs(end.X-start.X))
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
	path := FindPath(mp, Node{0, 5}, Node{9, 3})
	fmt.Printf("%+v\n", path)
	for y := 0; y < mp.height; y++ {
		for x := 0; x < mp.width; x++ {
			pth := false
			for _, p := range path {
				if p.X == x && p.Y == y {
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
