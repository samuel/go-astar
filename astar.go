package astar

import (
	"container/heap"
)

type Edge struct {
	Node int     // destination node
	Cost float64 // cost to move to the node
}

type Graph interface {
	Neighbors(node int) []Edge
	HeuristicCost(start, end int) float64
}

type nodeInfo struct {
	node          int
	parent        *nodeInfo // the node from which we came to get here
	count         int       // count lets use know how long the path is when we reach the end without traversing it
	cost          float64   // current cost from start node to this node
	predictedCost float64   // heuristic cost from this node to end node
	open          bool
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

func (nl *nodeList) RemoveNodeInfo(ni *nodeInfo) {
	for i, nodeInfo := range *nl {
		if nodeInfo.node == ni.node {
			heap.Remove(nl, i)
			break
		}
	}
}

func FindPath(mp Graph, start, end int) []int {
	nodes := make(map[int]*nodeInfo)
	// The open heap is ordered by the sum of current cost + heuristic cost
	nl := nodeList(make([]*nodeInfo, 0))
	open := &nl
	heap.Init(open)

	// Add the start node to the openlist
	ni := &nodeInfo{
		node:          start,
		parent:        nil,
		count:         1,
		cost:          0,
		predictedCost: mp.HeuristicCost(start, end),
		open:          true,
	}
	open.AddNodeInfo(ni)
	nodes[ni.node] = ni

	var path []int
	for {
		current := open.PopBest()
		if current.node == end {
			// If we reached the end node then we know the optimal path. Traverse
			// it (backwards) and return an array of node IDs.
			path = make([]int, current.count)
			for i, n := current.count-1, current; n != nil; i, n = i-1, n.parent {
				path[i] = n.node
			}
			break
		}
		current.open = false
		neighbors := mp.Neighbors(current.node)
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
				ni := &nodeInfo{
					node:          edge.Node,
					parent:        current,
					count:         current.count + 1,
					cost:          cost,
					predictedCost: mp.HeuristicCost(edge.Node, end),
					open:          true,
				}
				open.AddNodeInfo(ni)
				nodes[edge.Node] = ni
			} else if cost < ni.cost {
				// We've seen this node and the current path is cheaper
				// so update the changed info and add it to the open list
				// (replacing if necessary).
				if ni.open {
					open.RemoveNodeInfo(ni)
				}
				ni.open = true
				ni.parent = current
				ni.count = current.count + 1
				ni.cost = cost
				open.AddNodeInfo(ni)
			}
		}
	}
	return path
}
