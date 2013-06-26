package astar

import (
	"container/heap"
)

type Node struct {
	X int
	Y int
}

type Edge struct {
	Node Node
	Cost float64 // cost to move to this node
}

type nodeInfo struct {
	node   Node
	parent *nodeInfo
	count  int
	g      float64 // current cost from start node to this node
	h      float64 // heuristic cost from this node to end node
}

type Map interface {
	Neighbors(node Node) []Edge
	HeuristicCost(start Node, end Node) float64
}

type nodeList []*nodeInfo

func (nl *nodeList) Len() int {
	return len(*nl)
}

func (nl *nodeList) Less(i, j int) bool {
	l := *nl
	li := l[i]
	lj := l[j]
	return (li.g + li.h) < (lj.g + lj.h)
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

func (nl *nodeList) FindNode(n Node) (*nodeInfo, int) {
	for i, ni := range *nl {
		if ni.node == n {
			return ni, i
		}
	}
	return nil, -1
}

func (nl *nodeList) Remove(i int) {
	heap.Remove(nl, i)
}

func FindPath(mp Map, start Node, end Node) []Node {
	closed := make(map[Node]*nodeInfo)
	nl := nodeList(make([]*nodeInfo, 0))
	open := &nl
	heap.Init(open)
	open.Push(&nodeInfo{start, nil, 1, 0, mp.HeuristicCost(start, end)})

	var path []Node
	for {
		parent := open.PopBest()
		if parent.node == end {
			path = make([]Node, parent.count)
			for i, n := parent.count-1, parent; n != nil; i, n = i-1, n.parent {
				path[i] = n.node
			}
			break
		}
		closed[parent.node] = parent
		neighbors := mp.Neighbors(parent.node)
		for _, e := range neighbors {
			n := e.Node
			if n == parent.node {
				continue
			}

			cost := parent.g + e.Cost
			ni, nii := open.FindNode(n)
			if ni != nil && cost < ni.g {
				open.Remove(nii)
				ni = nil
			} else {
				ni = closed[n]
				if ni != nil && cost < ni.g {
					delete(closed, n)
					ni = nil
				}
			}
			if ni == nil {
				ni := &nodeInfo{
					node:   n,
					parent: parent,
					count:  parent.count + 1,
					g:      cost,
					h:      mp.HeuristicCost(n, end),
				}
				open.AddNodeInfo(ni)
			}
		}
	}
	return path
}
