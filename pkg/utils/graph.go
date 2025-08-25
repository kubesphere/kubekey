package utils

// Graph a Kahn graph,use for checking file import cycle
type Graph struct {
	edges    map[string][]string
	indegree map[string]int
}

// NewGraph return a graph struct for playbook cycle import check
func NewGraph() *Graph {
	return &Graph{
		edges:    make(map[string][]string),
		indegree: make(map[string]int),
	}
}

// AddEdgeAndCheckCycle add graph and check cycle
func (g *Graph) AddEdgeAndCheckCycle(a, b string) bool {
	g.edges[a] = append(g.edges[a], b)

	g.indegree[b]++
	if _, exists := g.indegree[a]; !exists {
		g.indegree[a] = 0
	}

	return g.HasCycle()
}

// HasCycle check current graph has cycle or not
func (g *Graph) HasCycle() bool {
	indegreeCopy := make(map[string]int)
	for k, v := range g.indegree {
		indegreeCopy[k] = v
	}

	queue := []string{}
	for node, deg := range indegreeCopy {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	count := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		count++

		for _, neighbor := range g.edges[node] {
			indegreeCopy[neighbor]--
			if indegreeCopy[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return count != len(indegreeCopy)
}
