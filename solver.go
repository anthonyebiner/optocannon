package main

import (
	graph2 "github.com/yourbasic/graph"
)

func (g *graph) greedy() ([]int, []edge, int, int) {
	gc := g.copy()
	numNodes := len(gc.nodes)
	var c, k, maxNodes, maxEdges int

	if numNodes > 50 {
		c, maxNodes = 5, 5
		k, maxEdges = 100, 100
	} else if numNodes > 30 {
		c, maxNodes = 3, 3
		k, maxEdges = 50, 50
	} else {
		c, maxNodes = 1, 1
		k, maxEdges = 15, 15
	}

	nodesToRemove := make([]int, 0)
	edgesToRemove := make([]edge, 0)

	for c > 0 || k > 0 {
		gc = g.copy()
		gc.removeNodesAndEdges(nodesToRemove, edgesToRemove)
		best, err := gc.shortestPath()
		CheckError(err)


		found := false
		if len(best.Path) > 2 && c > 0 {
			var nodeToRemove int
			var maxDist int64
			maxDist = 0
			for i := 1; i < len(best.Path) - 1; i++ {
				node := best.Path[i]
				nodesToRemove = append(nodesToRemove, node)
				gcc := g.copy()
				gcc.removeNodesAndEdges(nodesToRemove, edgesToRemove)
				connected := graph2.Connected(gcc.toSimpleGraph())
				best2, err := gcc.shortestPath()
				if err == nil && connected && best2.Distance > maxDist {
					found = true
					nodeToRemove = node
					maxDist = best2.Distance
				}
				nodesToRemove = nodesToRemove[0:len(nodesToRemove)-1]
			}
			if found {
				nodesToRemove = append(nodesToRemove, nodeToRemove)
				c--
			}
		}
		if !found && len(best.Path) > 1 && k > 0 {
			var edgeToRemove edge
			var maxDist int64
			maxDist = 0
			for i := 0; i < len(best.Path) - 1; i++ {
				edge, _ := gc.getEdge(best.Path[i], best.Path[i+1])
				edgesToRemove = append(edgesToRemove, edge)
				gcc := g.copy()
				gcc.removeNodesAndEdges(nodesToRemove, edgesToRemove)
				connected := graph2.Connected(gcc.toSimpleGraph())
				best2, err := gcc.shortestPath()
				if err != nil || !connected {
					edgesToRemove = edgesToRemove[0:len(edgesToRemove)-1]
					continue
				}
				if best2.Distance > maxDist {
					found = true
					edgeToRemove = edge
					maxDist = best2.Distance
				}
				edgesToRemove = edgesToRemove[0:len(edgesToRemove)-1]
			}

			if found {
				k--
				edgesToRemove = append(edgesToRemove, edgeToRemove)
			}
		}

		if !found {
			break
		}
	}

	if len(nodesToRemove) > maxNodes {
		panic("removing more nodes")
	}
	if len(edgesToRemove) > maxEdges {
		panic("removing more edges")
	}

	return nodesToRemove, edgesToRemove, maxNodes, maxEdges
}
