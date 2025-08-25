package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphHasCycle(t *testing.T) {
	testcases := []struct {
		name   string
		graph  KahnGraph
		except bool
	}{
		{
			// a -> b
			name: "Single head, single depth, no cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b"},
				},
				indegree: map[string]int{
					"a": 0,
					"b": 1,
				},
			},
			except: false,
		},
		{
			// a -> a
			name: "Single head, single depth, has cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"a"},
				},
				indegree: map[string]int{
					"a": 1,
				},
			},
			except: true,
		},
		{
			// a -> b
			// a -> c
			name: "Multiple heads, single depth, no cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b", "c"},
				},
				indegree: map[string]int{
					"a": 0,
					"b": 1,
					"c": 1,
				},
			},
			except: false,
		},
		{
			// a -> b
			// a -> a
			name: "Multiple heads, single depth, has cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b", "a"},
				},
				indegree: map[string]int{
					"a": 1,
					"b": 1,
				},
			},
			except: true,
		},
		{
			// a -> b -> c
			name: "Single head, multiple depth, no cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b"},
					"b": {"c"},
				},
				indegree: map[string]int{
					"a": 0,
					"b": 1,
					"c": 1,
				},
			},
			except: false,
		},
		{
			// a -> b -> a
			name: "Single head, multiple depth, has cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b"},
					"b": {"a"},
				},
				indegree: map[string]int{
					"a": 1,
					"b": 1,
				},
			},
			except: true,
		},
		{
			// a -> b
			// a -> c -> d
			// a -> d -> b
			name: "Multiple heads, multiple depth, no cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b", "c", "d"},
					"c": {"d"},
					"d": {"b"},
				},
				indegree: map[string]int{
					"a": 0,
					"b": 2,
					"c": 1,
					"d": 2,
				},
			},
			except: false,
		},
		{
			// a -> b
			// a -> c -> d
			// a -> d -> c
			name: "Multiple heads, multiple depth, has cycle",
			graph: KahnGraph{
				edges: map[string][]string{
					"a": {"b", "c", "d"},
					"c": {"d"},
					"d": {"c"},
				},
				indegree: map[string]int{
					"a": 0,
					"b": 1,
					"c": 2,
					"d": 2,
				},
			},
			except: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, tc.graph.hasCycle())
		})
	}
}
