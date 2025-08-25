package utils

import (
	"testing"
)

// TestNewGraph
// new empty graph
func TestNewGraph(t *testing.T) {
	graph := NewGraph()

	if graph.edges == nil {
		t.Error("edges map should be initialized")
	}

	if graph.indegree == nil {
		t.Error("indegree map should be initialized")
	}
}

// TestAddEdgeAndCheckCycle_NoCycle
// first check a -> b
// then make route a -> b -> c and check
func TestAddEdgeAndCheckCycle_NoCycle(t *testing.T) {
	graph := NewGraph()

	hasCycle := graph.AddEdgeAndCheckCycle("a", "b")
	if hasCycle {
		t.Error("shoud not check to cycle")
	}

	hasCycle = graph.AddEdgeAndCheckCycle("b", "c")
	if hasCycle {
		t.Error("shoud not check to cycle")
	}

	if len(graph.edges["a"]) != 1 || graph.edges["a"][0] != "b" {
		t.Error("a->b should add success")
	}

	if len(graph.edges["b"]) != 1 || graph.edges["b"][0] != "c" {
		t.Error("b->c should add success")
	}

	if graph.indegree["a"] != 0 {
		t.Error("in degree for a should be 0")
	}

	if graph.indegree["b"] != 1 {
		t.Error("in degree for b should be 1")
	}

	if graph.indegree["c"] != 1 {
		t.Error("in degree for c should be 1")
	}
}

// TestAddEdgeAndCheckCycle_WithCycle
// a -> b -> c -> a
// cause a cycle import
func TestAddEdgeAndCheckCycle_WithCycle(t *testing.T) {
	graph := NewGraph()

	graph.AddEdgeAndCheckCycle("a", "b")
	graph.AddEdgeAndCheckCycle("b", "c")
	hasCycle := graph.AddEdgeAndCheckCycle("c", "a")

	if !hasCycle {
		t.Error("should check to cycle")
	}
}

// TestAddEdgeAndCheckCycle_SelfLoop
// a -> a
// import self
func TestAddEdgeAndCheckCycle_SelfLoop(t *testing.T) {
	graph := NewGraph()

	hasCycle := graph.AddEdgeAndCheckCycle("a", "a")

	if !hasCycle {
		t.Error("self cycle should check as a cycle")
	}
}

// TestAddEdgeAndCheckCycle_MultiplePathsNoCycle
/*
     | -> b -> | d -> e
a -> |           ^
	 | -> c -> | |
*/
func TestAddEdgeAndCheckCycle_MultiplePathsNoCycle(t *testing.T) {
	graph := NewGraph()

	graph.AddEdgeAndCheckCycle("a", "b")
	graph.AddEdgeAndCheckCycle("a", "c")
	graph.AddEdgeAndCheckCycle("b", "d")
	graph.AddEdgeAndCheckCycle("c", "d")
	hasCycle := graph.AddEdgeAndCheckCycle("d", "e")

	if hasCycle {
		t.Error("should not check to cycle")
	}
}

// TestAddEdgeAndCheckCycle_MultipleCycles
// a -> b -> c ->a
// x -> y
// a graph with 2 start node, first node cause cycle
func TestAddEdgeAndCheckCycle_MultipleCycles(t *testing.T) {
	graph := NewGraph()

	graph.AddEdgeAndCheckCycle("a", "b")
	graph.AddEdgeAndCheckCycle("b", "c")
	graph.AddEdgeAndCheckCycle("c", "a")

	graph.AddEdgeAndCheckCycle("x", "y")
	hasCycle := graph.AddEdgeAndCheckCycle("y", "x")

	if !hasCycle {
		t.Error("should check to cycle")
	}
}

// TestHasCycle_EmptyGraph
// empty graph
func TestHasCycle_EmptyGraph(t *testing.T) {
	graph := NewGraph()

	if graph.HasCycle() {
		t.Error("empty graph should not have cycle")
	}
}

// TestHasCycle_SingleNode
// only one node named a
func TestHasCycle_SingleNode(t *testing.T) {
	graph := NewGraph()

	graph.indegree["a"] = 0

	if graph.HasCycle() {
		t.Error("single node graph should not have cycle")
	}
}

// TestHasCycle_AfterRemovingCycle
// graph = a -> b -> c -> a
// graph2 = a -> b -> c
func TestHasCycle_AfterRemovingCycle(t *testing.T) {
	graph := NewGraph()

	graph.AddEdgeAndCheckCycle("a", "b")
	graph.AddEdgeAndCheckCycle("b", "c")
	graph.AddEdgeAndCheckCycle("c", "a")

	if !graph.HasCycle() {
		t.Error("should check to cycle")
	}

	graph2 := NewGraph()
	graph2.AddEdgeAndCheckCycle("a", "b")
	graph2.AddEdgeAndCheckCycle("b", "c")

	if graph2.HasCycle() {
		t.Error("should not check to cycle")
	}
}

// TestAddEdgeAndCheckCycle_DuplicateEdges
// a -> b
// add route twice
func TestAddEdgeAndCheckCycle_DuplicateEdges(t *testing.T) {
	graph := NewGraph()

	hasCycle1 := graph.AddEdgeAndCheckCycle("a", "b")
	hasCycle2 := graph.AddEdgeAndCheckCycle("a", "b")

	if hasCycle1 || hasCycle2 {
		t.Error("multi side should not have cycle")
	}

	if len(graph.edges["a"]) != 2 || graph.edges["a"][0] != "b" || graph.edges["a"][1] != "b" {
		t.Error("multi side should import success")
	}

	if graph.indegree["b"] != 2 {
		t.Error("in degree for b should be 2")
	}
}

// TestAddEdgeAndCheckCycle_MultiplePathsNoCycle
/*
     | -> b -> | d -> e -> | -> f
a -> |           ^         | -> g
	 | -> c -> | |         | -> h
     | -> -> ->  |         | -> b
with a cycle b -> d -> e -> b
*/
func TestAddEdgeAndCheckCycle_MultiplePathsWithCycle(t *testing.T) {
	graph := NewGraph()

	graph.AddEdgeAndCheckCycle("a", "b")
	graph.AddEdgeAndCheckCycle("a", "c")
	graph.AddEdgeAndCheckCycle("b", "d")
	graph.AddEdgeAndCheckCycle("c", "d")
	graph.AddEdgeAndCheckCycle("d", "e")
	graph.AddEdgeAndCheckCycle("a", "d")
	graph.AddEdgeAndCheckCycle("e", "f")
	graph.AddEdgeAndCheckCycle("e", "g")
	graph.AddEdgeAndCheckCycle("e", "h")
	hasCycle := graph.AddEdgeAndCheckCycle("e", "b")

	if !hasCycle {
		t.Error("should check to cycle")
	}
}
