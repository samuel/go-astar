package astar

import (
	"math"
)

const (
	maxDefaultMapCapacity = 131072
	defaultListCapacity   = 4096
)

type nodeInfo struct {
	node          Node
	parent        Node    // the node from which we came to get here
	index         int     // index of the node in the heap
	cost          float32 // current cost from start node to this node
	predictedCost float32 // heuristic cost from this node to end node
}

type state struct {
	info    map[Node]*nodeInfo
	heap    []*nodeInfo
	maxCost float32
}

func (s *state) pathToNode(node *nodeInfo) []Node {
	path := make([]Node, 0, 128)
	for n := node; n != nil; n = s.info[n.parent] {
		path = append(path, n.node)
	}
	// Reverse the path since we built it backwards
	n := len(path) / 2
	for i := range n {
		j := len(path) - i - 1
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func newState(capacity int) *state {
	return &state{
		info:    make(map[Node]*nodeInfo, capacity),
		heap:    make([]*nodeInfo, 0, defaultListCapacity),
		maxCost: float32(math.Inf(1)),
	}
}

func (nl *state) less(i, j int) bool {
	li := nl.heap[i]
	lj := nl.heap[j]
	return (li.cost + li.predictedCost) < (lj.cost + lj.predictedCost)
}

func (nl *state) swap(i, j int) {
	l := nl.heap
	l[i], l[j] = l[j], l[i]
	l[i].index = i
	l[j].index = j
}

func (nl *state) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !nl.less(j, i) {
			break
		}
		nl.swap(i, j)
		j = i
	}
}

func (nl *state) down(i, n int) {
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

func (nl *state) popBest() *nodeInfo {
	n := len(nl.heap) - 1
	if n < 0 {
		return nil
	}
	nl.swap(0, n)
	nl.down(0, n)
	v := nl.heap[n]
	nl.heap = nl.heap[:n]
	v.index = -1
	return v
}

func (nl *state) addNodeInfo(ni *nodeInfo) {
	nl.info[ni.node] = ni
	nl.heap = append(nl.heap, ni)
	i := len(nl.heap) - 1
	ni.index = i
	nl.up(i)
}

func (nl *state) updateNodeInfo(ni *nodeInfo) {
	index := ni.index
	n := len(nl.heap)
	nl.down(index, n)
	nl.up(index)
}

// Find the optimal path through the graph from start to end and
// return the nodes in order for the path. If no path is found
// because it's impossible to reach end from start then return an error.
func FindPath(mp Graph, start, end Node) ([]Node, error) {
	mapCapacity := int(end - start)
	if mapCapacity < 0 {
		mapCapacity = -mapCapacity
	}
	if mapCapacity > maxDefaultMapCapacity {
		mapCapacity = maxDefaultMapCapacity
	}
	// The open list is ordered by the sum of current cost + heuristic cost
	state := newState(mapCapacity)
	// Add the start node to the openlist
	pCost, err := mp.HeuristicCost(start, end)
	if err != nil {
		return nil, err
	}
	state.addNodeInfo(&nodeInfo{
		node:          start,
		parent:        -1,
		cost:          0.0,
		predictedCost: float32(pCost),
	})

	edgeSlice := make([]Edge, 0, 8)
	for {
		current := state.popBest()
		if current == nil {
			return nil, ErrImpossible
		}
		if current.node == end {
			// If we reached the end node then we know the optimal path. Traverse
			// it (backwards) and return an array of node IDs.
			return state.pathToNode(current), nil
		}
		if current.cost >= state.maxCost {
			continue
		}
		if dbg, ok := mp.(Debug); ok {
			dbg.VisitedNode(current.node, current.parent, float64(current.cost), float64(current.predictedCost))
		}
		neighbors, err := mp.Neighbors(current.node, edgeSlice[:0])
		if err != nil {
			return nil, err
		}
		for _, edge := range neighbors {
			// Don't try go backwards
			if edge.Node == current.parent {
				continue
			}

			// Cost for the neighbor node is the current cost plus the
			// cost to get to that node.
			cost := current.cost + float32(edge.Cost)

			ni := state.info[edge.Node]
			if ni == nil {
				// We haven't seen this node so add it to the open list.
				pCost, err := mp.HeuristicCost(edge.Node, end)
				if err != nil {
					return nil, err
				}
				ni = &nodeInfo{
					node:          edge.Node,
					parent:        current.node,
					cost:          cost,
					predictedCost: float32(pCost),
				}
				state.addNodeInfo(ni)
			} else if cost < ni.cost {
				// We've seen this node and the current path is cheaper
				// so update the changed info and add it to the open list
				// (replacing if necessary).
				ni.parent = current.node
				ni.cost = cost
				if ni.index >= 0 {
					state.updateNodeInfo(ni)
				} else {
					state.addNodeInfo(ni)
				}
			} else if edge.Node == end {
				if cost < state.maxCost {
					state.maxCost = cost
				}
				if pp, ok := mp.(PossiblePath); ok {
					path := append(state.pathToNode(current), end)
					pp.PossiblePath(path, float64(cost))
				}
				ni = nil
			}
			if ni != nil && edge.Node == end {
				if cost < state.maxCost {
					state.maxCost = cost
				}
				if pp, ok := mp.(PossiblePath); ok {
					pp.PossiblePath(state.pathToNode(ni), float64(ni.cost))
				}
			}
		}
	}
}
