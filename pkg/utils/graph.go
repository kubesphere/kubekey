package utils

// KahnGraph represents a directed graph and provides efficient cycle detection using Kahn's algorithm.
// Kahn's algorithm repeatedly removes nodes with in-degree 0 (i.e., nodes with no dependencies).
// If a cycle exists, nodes in the cycle will never have in-degree 0, so the algorithm cannot process all nodes, thus detecting the presence of a cycle.
type KahnGraph struct {
	// edges is an adjacency list: each node stores all its outgoing edges (target nodes).
	// key: source node, value: slice of target nodes
	edges map[string][]string

	// indegree records the in-degree (number of incoming edges) for each node.
	// key: node name, value: in-degree count
	indegree map[string]int
}

// NewKahnGraph initializes and returns an empty KahnGraph.
func NewKahnGraph() *KahnGraph {
	return &KahnGraph{
		edges:    make(map[string][]string),
		indegree: make(map[string]int),
	}
}

// AddEdgeAndCheckCycle adds a directed edge from node a to node b and immediately checks if a cycle is formed.
// Parameters:
//   - a: source node
//   - b: target node
//
// Returns:
//   - true if adding the edge creates a cycle; false otherwise
func (g *KahnGraph) AddEdgeAndCheckCycle(a, b string) bool {
	// Add an edge from a to b
	g.edges[a] = append(g.edges[a], b)

	// Increment the in-degree of b (one more edge points to b)
	g.indegree[b]++

	// Ensure a is present in the in-degree map (initialize to 0 if not present)
	if _, exists := g.indegree[a]; !exists {
		g.indegree[a] = 0
	}

	// After adding the new edge, check if a cycle exists
	return g.hasCycle()
}

// hasCycle determines whether the current graph contains a cycle.
// Returns:
//   - true if a cycle exists; false otherwise
func (g *KahnGraph) hasCycle() bool {
	// Make a copy of the in-degree map to avoid modifying the original data
	indegreeCopy := make(map[string]int, len(g.indegree))
	for node, degree := range g.indegree {
		indegreeCopy[node] = degree
	}

	// Initialize a queue to collect all nodes with in-degree 0
	queue := make([]string, 0)
	for node, degree := range indegreeCopy {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	processed := 0 // Count of nodes processed by topological sort

	// Process all nodes with in-degree 0 and update their neighbors' in-degree
	for len(queue) > 0 {
		// Dequeue the first node
		node := queue[0]
		queue = queue[1:]
		processed++

		// For each neighbor, decrement its in-degree
		for _, neighbor := range g.edges[node] {
			indegreeCopy[neighbor]--
			// If neighbor's in-degree becomes 0, add it to the queue
			if indegreeCopy[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If all nodes are processed, there is no cycle; otherwise, a cycle exists
	return processed != len(indegreeCopy)
}
