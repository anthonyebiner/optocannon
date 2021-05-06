package main

import "fmt"

func main() {
	db := connect()

	PassNum := 25
	MaxIterations := 3000000
	MaxNumResets := 5
	MaxTemp := 500.0
	MinTemp := 5.0

	for {
		fmt.Println("---")
		fmt.Println()

		g, _ := grabSmallest(db, PassNum)
		_, err := g.shortestPath()
		CheckError(err)

		sol := startingSolution(g)
		sol.anneal(MaxIterations, MaxNumResets, MaxTemp, MinTemp)
		sol.optimize()

		addSolution(db, g.id, g, sol.nodesToRemove, sol.edgesToRemove, false, PassNum)
	}

	db.Close()
}