package astar

import (
	"errors"
)

var ErrImpossible = errors.New("astar: no path exists between start and end")

type Edge struct {
	Node int64   // destination node
	Cost float64 // cost to move to the node
}

type Graph interface {
	// Edges is passed in for reuse. This method gets called a large number of times
	// so we don't want to allocate an Edge slice for every call.
	Neighbors(node int64, edges []Edge) ([]Edge, error)
	HeuristicCost(start, end int64) (float64, error)
}

// If a graph implementation implements the PossiblePath interface then
// it can receive intermediate results before the algorithms converges on
// an optimal path.
type PossiblePath interface {
	PossiblePath(path []int64, cost float64)
}

type Debug interface {
	VisitedNode(node, parentNode int64, currentCost, predictedCost float64)
}
