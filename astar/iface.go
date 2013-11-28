package astar

import (
	"errors"
)

var ErrImpossible = errors.New("astar: no path exists between start and end")

type Node int64

type Edge struct {
	Node Node    // destination node
	Cost float64 // cost to move to the node
}

type Graph interface {
	// Edges is passed in for reuse. This method gets called a large number of times
	// so we don't want to allocate an Edge slice for every call.
	Neighbors(node Node, edges []Edge) ([]Edge, error)
	HeuristicCost(start, end Node) (float64, error)
}

// If a graph implementation implements the PossiblePath interface then
// it can receive intermediate results before the algorithms converges on
// an optimal path.
type PossiblePath interface {
	PossiblePath(path []Node, cost float64)
}

type Debug interface {
	VisitedNode(node, parentNode Node, currentCost, predictedCost float64)
}
