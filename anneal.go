package main

import (
	"fmt"
	graph2 "github.com/yourbasic/graph"
	"math"
	"math/rand"
	"time"
)

type solution struct {
	graphOrig *graph
	nodesToRemove []int
	edgesToRemove []edge
	maxNodesToRemove int
	maxEdgesToRemove int
	energy int64
}

func (s *solution) copyEdges(edges []edge) {
	s.edgesToRemove = nil
	s.edgesToRemove = append(s.edgesToRemove, edges...)
}

func (s *solution) copyNodes(nodes []int) {
	s.nodesToRemove = nil
	s.nodesToRemove = append(s.nodesToRemove, nodes...)
}

func (s *solution) copy() *solution{
	sc := solution{
		graphOrig: s.graphOrig.copy(),
		maxNodesToRemove: s.maxNodesToRemove,
		maxEdgesToRemove: s.maxEdgesToRemove,
		nodesToRemove: make([]int, len(s.nodesToRemove)),
		edgesToRemove: make([]edge, len(s.edgesToRemove)),
		energy: s.energy,
	}

	sc.copyNodes(s.nodesToRemove)
	sc.copyEdges(s.edgesToRemove)

	return &sc
}

func (s *solution) copyFrom(s2 *solution) {
	s.copyNodes(s2.nodesToRemove)
	s.copyEdges(s2.edgesToRemove)
	s.energy = s2.energy
}

func (s *solution) containsEdge(e edge) bool  {
	for _, en := range s.edgesToRemove {
		if (en.from == e.from && en.to == e.to) || (en.from == e.to && en.to == e.from) {
			return true
		}
	}
	return false
}

func (s *solution) containsNode(n int) bool {
	for _, i := range s.nodesToRemove {
		if i == n {
			return true
		}
	}
	return false
}

func (s *solution) calcEnergy() int64 {
	gc := s.graphOrig.copy()
	gc.removeNodesAndEdges(s.nodesToRemove, s.edgesToRemove)
	sg := gc.toSimpleGraph()
	if !graph2.Connected(sg) {
		return -1000000000
	}
	best, err := gc.shortestPath()
	if err != nil {
		return -1000000000
	}
	return best.Distance
}

func (s *solution) move() {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(100)

	if r > 98 && len(s.nodesToRemove) > 0 { // Add one more node
		nodeR := rand.Intn(len(s.nodesToRemove))
		s.nodesToRemove = append(s.nodesToRemove[:nodeR], s.nodesToRemove[nodeR+1:]...)

	} else if r > 96 && len(s.edgesToRemove) > 0 { // Add one more edge
		edgeR := rand.Intn(len(s.edgesToRemove))
		s.edgesToRemove = append(s.edgesToRemove[:edgeR], s.edgesToRemove[edgeR+1:]...)

	} else if r > 92 && len(s.nodesToRemove) < s.maxNodesToRemove { // Remove one more node
		nodeA := s.graphOrig.nodes[rand.Intn(len(s.graphOrig.nodes)-2)+1]

		if !s.containsNode(nodeA) {
			s.nodesToRemove = append(s.nodesToRemove, nodeA)
		}

	} else if r > 88 && len(s.edgesToRemove) < s.maxEdgesToRemove  { // Remove one more edge
		edgeA := s.graphOrig.edges[rand.Intn(len(s.graphOrig.edges))]

		if !s.containsEdge(edgeA) {
			s.edgesToRemove = append(s.edgesToRemove, edgeA)
		}

	} else if r > 15 && len(s.edgesToRemove) > 0 { // Swap edges
		edgeR := rand.Intn(len(s.edgesToRemove))
		edgeA := s.graphOrig.edges[rand.Intn(len(s.graphOrig.edges))]

		if !s.containsEdge(edgeA) {
			s.edgesToRemove[edgeR] = edgeA
		}

	} else if len(s.nodesToRemove) > 0 { // Swap nodes
		nodeR := rand.Intn(len(s.nodesToRemove))
		nodeA := s.graphOrig.nodes[rand.Intn(len(s.graphOrig.nodes)-2)+1]

		if !s.containsNode(nodeA) {
			s.nodesToRemove[nodeR] = nodeA
		}
	}

	s.energy = s.calcEnergy()
}

func startingSolution(graphOrig graph) *solution {
	gc := graphOrig.copy()
	nodesToRemove, edgesToRemove, maxNodesToRemove, maxEdgesToRemove := gc.greedy()
	s := solution{
		graphOrig: graphOrig.copy(),
		maxNodesToRemove: maxNodesToRemove,
		maxEdgesToRemove: maxEdgesToRemove,
	}
	s.copyEdges(edgesToRemove)
	s.copyNodes(nodesToRemove)
	s.energy = s.calcEnergy()
	return &s
}

func (s *solution) anneal(maxIterations int, maxDriftFactor int, maxTemp float64, minTemp float64) {
	prevSol := s.copy()
	resetSol := s.copy()

	accepts, improves, resets := 0, 0, 0

	maxDriftFromGlobalMinimum := maxIterations / maxDriftFactor

	tempFactor := -math.Log(maxTemp / minTemp)
	countSinceReset := maxDriftFromGlobalMinimum

	fmt.Println(" Temperature        Energy    Accept   Improve      Resets    Nodes   Edges")


	for i := 1; i < maxIterations; i++ {
		temp := maxTemp * math.Exp(tempFactor*float64(i) / float64(maxIterations))
		//x := float64(i/(maxIterations/1223))
		//exp := math.Exp(-0.004*x)
		//temp := ((exp*math.Sin(x/(4.0*math.Pi)+math.Pi/2.0)+exp)*500)*(maxTemp/1000.0) + minTemp

		s.move()
		dE := -s.energy + prevSol.energy


		if dE > 0 &&  math.Exp(float64(-dE) / temp) < rand.Float64() {
			if countSinceReset < 0 { // RESET
				resets++
				s.copyFrom(resetSol)
				prevSol.copyFrom(resetSol)
				countSinceReset = maxDriftFromGlobalMinimum
			} else { // DENIED
				s.copyFrom(prevSol)
			}
		} else { // ACCEPTED
			accepts++
			prevSol.copyFrom(s)

			if dE < 0.0 {
				improves++
			}

			if s.energy > resetSol.energy {
				resetSol.copyFrom(s)
				countSinceReset = maxDriftFromGlobalMinimum
			}
		}

		if i % 2500 == 0 {
			fmt.Print("\033[G\033[K")
			fmt.Println(fmt.Sprintf("%12.5f  %12d  %7.3f  %7.3f   %7d    %5d    %5d", temp, s.energy, float64(accepts) / float64(i), float64(improves) / float64(i), resets, len(s.nodesToRemove), len(s.edgesToRemove)))
			fmt.Print("\033[A")
		}

		countSinceReset--
	}
	s.copyFrom(resetSol)

	fmt.Println()
}

func (s *solution) optimize() {
	fmt.Println("optimizing")
	gc := s.graphOrig.copy()
	starting_dist, err := gc.shortestPath()
	CheckError(err)
	if len(s.edgesToRemove) == 0 {
		return
	}
	for true {
		gc = s.graphOrig.copy()
		gc.removeNodesAndEdges(s.nodesToRemove, s.edgesToRemove)
		best, err := gc.shortestPath()
		CheckError(err)

		var minDist int64
		minDist = 1000000
		var bestEdge int
		var be edge
		for i, edge := range s.edgesToRemove {
			s.edgesToRemove = append(s.edgesToRemove[0:i], s.edgesToRemove[i+1:]...)
			subG := s.graphOrig.copy()
			subG.removeNodesAndEdges(s.nodesToRemove, s.edgesToRemove)
			nextBest, err := subG.shortestPath()
			if err != nil {
				s.edgesToRemove = append(s.edgesToRemove[0:i+1], s.edgesToRemove[i:]...)
				s.edgesToRemove[i] = edge
				continue
			}
			if best.Distance - nextBest.Distance < minDist {
				minDist = best.Distance - nextBest.Distance
				bestEdge = i
				be = edge
			}
			s.edgesToRemove = append(s.edgesToRemove[0:i+1], s.edgesToRemove[i:]...)
			s.edgesToRemove[i] = edge
		}

		s.edgesToRemove = append(s.edgesToRemove[0:bestEdge], s.edgesToRemove[bestEdge+1:]...)
		gc = s.graphOrig.copy()
		gc.removeNodesAndEdges(s.nodesToRemove, s.edgesToRemove)
		best2, err := gc.shortestPath()
		CheckError(err)

		found := false
		var edgeToRemove edge
		if len(best2.Path) > 1 {
			var maxDist int64
			maxDist = 0
			for i := 0; i < len(best2.Path)-1; i++ {
				edge, _ := gc.getEdge(best2.Path[i], best2.Path[i+1])
				s.edgesToRemove = append(s.edgesToRemove, edge)
				gcc := s.graphOrig.copy()
				gcc.removeNodesAndEdges(s.nodesToRemove, s.edgesToRemove)
				connected := graph2.Connected(gcc.toSimpleGraph())
				best3, err := gcc.shortestPath()
				if err != nil || !connected {
					s.edgesToRemove = s.edgesToRemove[0 : len(s.edgesToRemove)-1]
					continue
				}
				if best3.Distance > maxDist && best3.Distance > best.Distance {
					found = true
					edgeToRemove = edge
					maxDist = best3.Distance
				}
				s.edgesToRemove = s.edgesToRemove[0 : len(s.edgesToRemove)-1]
			}
		}
		if !found {
			s.edgesToRemove = append(s.edgesToRemove, be)
			fmt.Println("no better edge found")
			break
		}
		s.edgesToRemove = append(s.edgesToRemove, edgeToRemove)

		gc = s.graphOrig.copy()
		gc.removeNodesAndEdges(s.nodesToRemove, s.edgesToRemove)
		best4, err := gc.shortestPath()
		CheckError(err)
		fmt.Println("optimized from ", best.Distance - starting_dist.Distance, " to ", best4.Distance - starting_dist.Distance)
	}
}