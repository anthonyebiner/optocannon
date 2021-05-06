package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/RyanCarrier/dijkstra"
	_ "github.com/lib/pq"
	simg "github.com/yourbasic/graph"
	"os"
	"strconv"
	"strings"
)

const (
	host = "0.0.0.0"
	dbname = "postgres"
	user = "yourusernamehere"
	password = "yourpasswordhere"
	port = 5433
)

type edge struct {
	from int
	to   int
	weight float64
}

type graph struct {
	nodes  []int
	edges  []edge
	start  int
	end    int
	id     int
}

func (g *graph) copy() *graph {
	gr := graph{
		start: g.start,
		end:   g.end,
		id:    g.id,
	}

	gr.nodes = append(gr.nodes, g.nodes...)
	gr.edges = append(gr.edges, g.edges...)
	return &gr
}

func (g *graph) getEdge(from int, to int) (edge, bool) {
	for _, e := range g.edges {
		if (e.from == from && e.to == to)  || (e.from == to && e.to == from) {
			return e, true
		}
	}
	return edge{}, false
}

func (g *graph) removeEdge(from int, to int) {
	for i := len(g.edges) - 1; i >= 0; i-- {
		e := g.edges[i]
		if (e.from == from && e.to == to)  || (e.from == to && e.to == from) {
			g.edges = append(g.edges[:i], g.edges[i+1:]...)
		}
	}
}

func (g *graph) removeNodesAndEdges(nodes []int, edges []edge) {
	for _, e := range edges {
		g.removeEdge(e.from, e.to)
	}
	for _, n := range nodes {
		g.removeNode(n)
	}
}

func (g *graph) removeNode(num int) {
	for i := len(g.edges) - 1; i >= 0; i-- {
		e := g.edges[i]
		if num == e.from || num == e.to {
			g.edges = append(g.edges[:i], g.edges[i+1:]...)
		}
	}

	for i := len(g.nodes) - 1; i >= 0; i-- {
		if num == g.nodes[i] {
			g.nodes = append(g.nodes[:i], g.nodes[i+1:]...)
			break
		}
	}
}

func (g *graph) toDijkstraGraph() dijkstra.Graph {
	dg := dijkstra.NewGraph()
	for _, num := range g.nodes {
		dg.AddVertex(num)
	}
	for _, e := range g.edges {
		err := dg.AddArc(e.from, e.to, int64(e.weight*1000))
		CheckError(err)
		err = dg.AddArc(e.to, e.from, int64(e.weight*1000))
		CheckError(err)
	}
	return *dg
}

func (g *graph) toSimpleGraph() *simg.Mutable {
	mapping := make(map[int]int)
	for i, n := range g.nodes {
		mapping[n] = i
	}

	sg := simg.New(len(g.nodes))
	for _, e := range g.edges {
		sg.AddBothCost(mapping[e.from], mapping[e.to], int64(e.weight*1000))
	}
	return sg
}

func (g *graph) calculateScore(nodes []int, edges []edge) int64 {
	gc := g.copy()
	bestO, err := gc.shortestPath()
	CheckError(err)
	gc.removeNodesAndEdges(nodes, edges)
	bestA, err := gc.shortestPath()
	CheckError(err)
	return bestA.Distance-bestO.Distance
}

func (g *graph) shortestPath() (dijkstra.BestPath, error) {
	dg := g.toDijkstraGraph()
	best, err := dg.Shortest(g.start, g.end)
	return best, err
}

func connect() *sql.DB {
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	return db
}

func grabSmallest(db *sql.DB, pass int) (graph, bool) {
	fmt.Println("Getting smallest graph")
	query := `SELECT id, input, size, output FROM graphs
				WHERE processing = FALSE AND solved = FALSE AND pass < $1
				ORDER BY edges ASC LIMIT 1;`

	rows, err := db.Query(query, pass)
	CheckError(err)
	defer rows.Close()

	if !rows.Next() {
		return graph{}, false
	}

	var id int
	var input string
	var size int
	var output string

	err = rows.Scan(&id, &input, &size, &output)
	CheckError(err)

	_, err = db.Exec(`UPDATE graphs SET processing = TRUE WHERE id = $1`, id)
	CheckError(err)

	g := parseResponse(id, input)

	return g, true
}

func addSolution(db *sql.DB, id int, ogGraph graph, nodes []int, edges []edge, solved bool, passNum int) {
	score := ogGraph.calculateScore(nodes, edges)
	fmt.Println("Solved graph " + strconv.Itoa(id) + ": " + strconv.Itoa(int(score)))

	output := ""
	output += strconv.Itoa(len(nodes)) + "\n"
	for _, city := range nodes {
		output += string(strconv.Itoa(city)) + "\n"
	}
	output += strconv.Itoa(len(edges)) + "\n"
	for _, road := range edges {
		output += strconv.Itoa(road.from) + " " + strconv.Itoa(road.to) + "\n"
	}

	_, err := db.Exec(`UPDATE graphs SET output = $1, solved = $2, score = $3, bestpass = $6 WHERE id = $4 AND (score IS NULL OR SCORE < $5)`,
		output, solved, int(score), id, int(score), passNum)
	CheckError(err)
	_, err = db.Exec(`UPDATE graphs SET pass = $1, processing = FALSE WHERE id = $2`,
		passNum, id)
	CheckError(err)
}

func parseResponse(id int, input string) graph {
	fmt.Println("Parsing response")
	g := graph{
		nodes: make([]int, 0),
		edges: make([]edge, 0),
		start: 0,
		id:    id,
	}
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		stringArr := strings.Fields(scanner.Text())
		if len(stringArr) == 1 {
			numVertices, err := strconv.Atoi(stringArr[0])
			CheckError(err)
			for i := 0; i < numVertices; i++ {
				g.nodes = append(g.nodes, i)
			}
			g.end = numVertices - 1
			continue
		}
		from, err := strconv.Atoi(stringArr[0])
		CheckError(err)
		to, err := strconv.Atoi(stringArr[1])
		CheckError(err)
		weight, err := strconv.ParseFloat(stringArr[2], 32)
		CheckError(err)

		e := edge{
			from:   from,
			to:     to,
			weight: weight,
		}
		edgeList := append(g.edges, e)
		g.edges = edgeList
	}
	return g
}

func export(db *sql.DB) {
	query := `SELECT name, output FROM graphs`

	rows, err := db.Query(query)
	CheckError(err)
	defer rows.Close()

	var name string
	var output string

	for rows.Next() {
		rows.Scan(&name, &output)
		dir := strings.Split(name, "-")[0]
		filename := "./output/" + dir + "/" + name + ".out"
		fmt.Println(filename)
		file, err := os.Create(filename)
		CheckError(err)
		_, err = file.WriteString(output)
		CheckError(err)
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

