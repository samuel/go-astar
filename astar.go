package astar

import (
	"errors"
)

const (
	maxDefaultMapCapacity = 131072
	defaultListCapacity   = 4096
)

var ErrImpossible = errors.New("astar: no path exists between start and end")

type Edge struct {
	Node int     // destination node
	Cost float64 // cost to move to the node
}

type Graph interface {
	// Edges is passed in for reuse. This method gets called a large number of times
	// so we don't want to allocate an Edge slice for every call.
	Neighbors(node int, edges []Edge) ([]Edge, error)
	HeuristicCost(start, end int) (float64, error)
}

type nodeInfo struct {
	node          int
	parent        *nodeInfo // the node from which we came to get here
	count         int       // count lets use know how long the path is when we reach the end without traversing it
	cost          float64   // current cost from start node to this node
	predictedCost float64   // heuristic cost from this node to end node
	index         int       // index of the node in the heap
	open          bool
}

type nodeList struct {
	nodes []*nodeInfo
}

func newNodeList() *nodeList {
	return &nodeList{
		nodes: make([]*nodeInfo, 0, defaultListCapacity),
	}
}

func (nl *nodeList) less(i, j int) bool {
	li := nl.nodes[i]
	lj := nl.nodes[j]
	return (li.cost + li.predictedCost) < (lj.cost + lj.predictedCost)
}

func (nl *nodeList) swap(i, j int) {
	l := nl.nodes
	l[i], l[j] = l[j], l[i]
	l[i].index = i
	l[j].index = j
}

func (nl *nodeList) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !nl.less(j, i) {
			break
		}
		nl.swap(i, j)
		j = i
	}
}

func (nl *nodeList) down(i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !nl.less(j1, j2) {
			j = j2 // = 2*i + 2  // right child
		}
		if !nl.less(j, i) {
			break
		}
		nl.swap(i, j)
		i = j
	}
}

func (nl *nodeList) PopBest() *nodeInfo {
	n := len(nl.nodes) - 1
	if n < 0 {
		return nil
	}
	nl.swap(0, n)
	nl.down(0, n)
	v := nl.nodes[n]
	nl.nodes = nl.nodes[:n]
	v.index = -1
	return v
}

func (nl *nodeList) AddNodeInfo(ni *nodeInfo) {
	nl.nodes = append(nl.nodes, ni)
	ni.index = len(nl.nodes) - 1
	nl.up(len(nl.nodes) - 1)
}

func (nl *nodeList) UpdateNodeInfo(ni *nodeInfo) {
	index := ni.index
	n := len(nl.nodes)
	nl.down(index, n)
	nl.up(index)
}

// Find the optimal path through the graph from start to end and
// return the nodes in order for the path. If no path is found
// because it's impossible to reach end from start then return an error.
func FindPath(mp Graph, start, end int) ([]int, error) {
	mapCapacity := end - start
	if mapCapacity < 0 {
		mapCapacity = -mapCapacity
	}
	if mapCapacity > maxDefaultMapCapacity {
		mapCapacity = maxDefaultMapCapacity
	}
	nodes := make(map[int]*nodeInfo, mapCapacity)
	// The open list is ordered by the sum of current cost + heuristic cost
	open := newNodeList()
	// Add the start node to the openlist
	pCost, err := mp.HeuristicCost(start, end)
	if err != nil {
		return nil, err
	}
	ni := &nodeInfo{
		node:          start,
		parent:        nil,
		count:         1,
		cost:          0,
		predictedCost: pCost,
		open:          true,
	}
	open.AddNodeInfo(ni)
	nodes[ni.node] = ni

	edgeSlice := make([]Edge, 0, 8)
	for {
		current := open.PopBest()
		if current == nil {
			return nil, ErrImpossible
		}
		if current.node == end {
			// If we reached the end node then we know the optimal path. Traverse
			// it (backwards) and return an array of node IDs.
			path := make([]int, current.count)
			for i, n := current.count-1, current; n != nil; i, n = i-1, n.parent {
				path[i] = n.node
			}
			return path, nil
		}
		current.open = false
		neighbors, err := mp.Neighbors(current.node, edgeSlice[:0])
		if err != nil {
			return nil, err
		}
		for _, edge := range neighbors {
			// Don't try go backwards
			if current.parent != nil && edge.Node == current.parent.node {
				continue
			}

			// Cost for the neighbor node is the current cost plus the
			// cost to get to that node.
			cost := current.cost + edge.Cost

			if ni := nodes[edge.Node]; ni == nil {
				// We haven't seen this node so add it to the open list.
				pCost, err := mp.HeuristicCost(edge.Node, end)
				if err != nil {
					return nil, err
				}
				ni := &nodeInfo{
					node:          edge.Node,
					parent:        current,
					count:         current.count + 1,
					cost:          cost,
					predictedCost: pCost,
					open:          true,
				}
				open.AddNodeInfo(ni)
				nodes[edge.Node] = ni
			} else if cost < ni.cost {
				// We've seen this node and the current path is cheaper
				// so update the changed info and add it to the open list
				// (replacing if necessary).
				wasOpen := ni.open
				ni.open = true
				ni.parent = current
				ni.count = current.count + 1
				ni.cost = cost
				if wasOpen {
					open.UpdateNodeInfo(ni)
				} else {
					open.AddNodeInfo(ni)
				}
			}
		}
	}
}
