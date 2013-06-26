package astar

import (
	"container/heap"
)

type Edge struct {
	Node int     // destination node
	Cost float64 // cost to move to the node
}

type Map interface {
	Neighbors(node int) []Edge
	HeuristicCost(start int, end int) float64
}

type nodeInfo struct {
	node          int
	parent        *nodeInfo // the node from which we came to get here
	count         int       // count lets use know how long the path is when we reach the end without traversing it
	cost          float64   // current cost from start node to this node
	predictedCost float64   // heuristic cost from this node to end node
}

type nodeList []*nodeInfo

func (nl *nodeList) Len() int {
	return len(*nl)
}

func (nl *nodeList) Less(i, j int) bool {
	l := *nl
	li := l[i]
	lj := l[j]
	return (li.cost + li.predictedCost) < (lj.cost + lj.predictedCost)
}

func (nl *nodeList) Swap(i, j int) {
	l := *nl
	l[i], l[j] = l[j], l[i]
}

func (nl *nodeList) Push(x interface{}) {
	*nl = append(*nl, x.(*nodeInfo))
}

func (nl *nodeList) Pop() interface{} {
	n := len(*nl)
	if n == 0 {
		return nil
	}
	v := (*nl)[n-1]
	*nl = (*nl)[:n-1]
	return v
}

func (nl *nodeList) Clear() {
	*nl = (*nl)[:0]
}

func (nl *nodeList) PopBest() *nodeInfo {
	return heap.Pop(nl).(*nodeInfo)
}

func (nl *nodeList) AddNodeInfo(ni *nodeInfo) {
	heap.Push(nl, ni)
}

// Find node in the list and return the *nodeInfo and index in the list.
// If the node is not in the list, then return nil, -1
func (nl *nodeList) FindNode(n int) (*nodeInfo, int) {
	for i, ni := range *nl {
		if ni.node == n {
			return ni, i
		}
	}
	return nil, -1
}

// Remove item at 'index'
func (nl *nodeList) Remove(index int) {
	heap.Remove(nl, index)
}

func FindPath(mp Map, start int, end int) []int {
	closed := make(map[int]*nodeInfo)
	// The open heap is ordered by the sum of current cost + heuristic cost
	nl := nodeList(make([]*nodeInfo, 0))
	open := &nl
	heap.Init(open)
	open.Push(&nodeInfo{start, nil, 1, 0, mp.HeuristicCost(start, end)})

	var path []int
	for {
		parent := open.PopBest()
		if parent.node == end {
			// If we reached the end node then we know the optimal path. Traverse
			// it (backwards) and return an array of node IDs.
			path = make([]int, parent.count)
			for i, n := parent.count-1, parent; n != nil; i, n = i-1, n.parent {
				path[i] = n.node
			}
			break
		}
		closed[parent.node] = parent
		neighbors := mp.Neighbors(parent.node)
		for _, e := range neighbors {
			n := e.Node

			// The node must not be a neighbor of itself
			if n == parent.node {
				continue
			}

			// Cost for the neighbor node is the current cost plus the
			// cost to get to that node.
			cost := parent.cost + e.Cost

			ni, nii := open.FindNode(n)
			if ni != nil && cost < ni.cost {
				open.Remove(nii)
				ni = nil
			} else {
				ni = closed[n]
				if ni != nil && cost < ni.cost {
					delete(closed, n)
					ni = nil
				}
			}
			if ni == nil {
				ni := &nodeInfo{
					node:          n,
					parent:        parent,
					count:         parent.count + 1,
					cost:          cost,
					predictedCost: mp.HeuristicCost(n, end),
				}
				open.AddNodeInfo(ni)
			}
		}
	}
	return path
}
