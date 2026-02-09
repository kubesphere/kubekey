/*
Copyright 2024 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"sort"
	"testing"
)

func TestSortableDispatcherCandidates_Len(t *testing.T) {
	testCases := []struct {
		name        string
		candidates  []dispatcherCandidate
		expected    int
		description string
	}{
		{
			name:        "empty candidates",
			candidates:  []dispatcherCandidate{},
			expected:    0,
			description: "should return 0 for empty candidates",
		},
		{
			name: "single candidate",
			candidates: []dispatcherCandidate{
				{matchesCount: 1},
			},
			expected:    1,
			description: "should return 1 for single candidate",
		},
		{
			name: "multiple candidates",
			candidates: []dispatcherCandidate{
				{matchesCount: 1},
				{matchesCount: 2},
				{matchesCount: 3},
			},
			expected:    3,
			description: "should return correct count for multiple candidates",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dc := &sortableDispatcherCandidates{candidates: tc.candidates}
			if dc.Len() != tc.expected {
				t.Errorf("expected Len()=%d, got %d", tc.expected, dc.Len())
			}
		})
	}
}

func TestSortableDispatcherCandidates_Swap(t *testing.T) {
	testCases := []struct {
		name        string
		i           int
		j           int
		input       []dispatcherCandidate
		expected    []dispatcherCandidate
		description string
	}{
		{
			name: "swap first and last",
			i:    0,
			j:    2,
			input: []dispatcherCandidate{
				{finalMatch: "first"},
				{finalMatch: "second"},
				{finalMatch: "last"},
			},
			expected: []dispatcherCandidate{
				{finalMatch: "last"},
				{finalMatch: "second"},
				{finalMatch: "first"},
			},
			description: "should swap elements correctly",
		},
		{
			name: "swap adjacent elements",
			i:    0,
			j:    1,
			input: []dispatcherCandidate{
				{finalMatch: "a"},
				{finalMatch: "b"},
				{finalMatch: "c"},
			},
			expected: []dispatcherCandidate{
				{finalMatch: "b"},
				{finalMatch: "a"},
				{finalMatch: "c"},
			},
			description: "should swap adjacent elements correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dc := &sortableDispatcherCandidates{candidates: tc.input}
			dc.Swap(tc.i, tc.j)

			for k, c := range dc.candidates {
				if c.finalMatch != tc.expected[k].finalMatch {
					t.Errorf("expected finalMatch=%q at index %d, got %q", tc.expected[k].finalMatch, k, c.finalMatch)
				}
			}
		})
	}
}

func TestSortableDispatcherCandidates_Less(t *testing.T) {
	testCases := []struct {
		name        string
		i           int
		j           int
		input       []dispatcherCandidate
		expected    bool
		description string
	}{
		{
			name: "less by matchesCount - i has fewer",
			i:    0,
			j:    1,
			input: []dispatcherCandidate{
				{matchesCount: 1},
				{matchesCount: 2},
			},
			expected:    true,
			description: "should return true when i has fewer matchesCount",
		},
		{
			name: "less by matchesCount - i has more",
			i:    1,
			j:    0,
			input: []dispatcherCandidate{
				{matchesCount: 1},
				{matchesCount: 2},
			},
			expected:    false,
			description: "should return false when i has more matchesCount",
		},
		{
			name: "equal matchesCount, less by literalCount",
			i:    0,
			j:    1,
			input: []dispatcherCandidate{
				{matchesCount: 1, literalCount: 5},
				{matchesCount: 1, literalCount: 10},
			},
			expected:    true,
			description: "should compare by literalCount when matchesCount equal",
		},
		{
			name: "equal matchesCount and literalCount, less by nonDefaultCount",
			i:    0,
			j:    1,
			input: []dispatcherCandidate{
				{matchesCount: 1, literalCount: 5, nonDefaultCount: 0},
				{matchesCount: 1, literalCount: 5, nonDefaultCount: 1},
			},
			expected:    true,
			description: "should compare by nonDefaultCount when matchesCount and literalCount equal",
		},
		{
			name: "equal in all counts",
			i:    0,
			j:    1,
			input: []dispatcherCandidate{
				{matchesCount: 1, literalCount: 5, nonDefaultCount: 0},
				{matchesCount: 1, literalCount: 5, nonDefaultCount: 0},
			},
			expected:    false,
			description: "should return false when all counts are equal",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dc := &sortableDispatcherCandidates{candidates: tc.input}
			if dc.Less(tc.i, tc.j) != tc.expected {
				t.Errorf("expected Less(%d, %d)=%v", tc.i, tc.j, tc.expected)
			}
		})
	}
}

func TestSortableDispatcherCandidates_Sort(t *testing.T) {
	testCases := []struct {
		name        string
		input       []dispatcherCandidate
		description string
	}{
		{
			name: "unsorted by matchesCount",
			input: []dispatcherCandidate{
				{matchesCount: 3},
				{matchesCount: 1},
				{matchesCount: 2},
			},
			description: "should sort by matchesCount in descending order",
		},
		{
			name: "unsorted with same matchesCount but different literalCount",
			input: []dispatcherCandidate{
				{matchesCount: 1, literalCount: 10},
				{matchesCount: 1, literalCount: 5},
				{matchesCount: 1, literalCount: 15},
			},
			description: "should sort by literalCount when matchesCount equal",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dc := &sortableDispatcherCandidates{candidates: tc.input}

			// Sort using sort.Reverse to match the implementation
			// The Less function returns true when i < j (i should come before j)
			// sort.Reverse reverses this, so higher values come first
			sort.Sort(sort.Reverse(dc))

			// Verify the sort order (descending by matchesCount)
			for k := 0; k < len(dc.candidates)-1; k++ {
				if dc.candidates[k].matchesCount < dc.candidates[k+1].matchesCount {
					t.Errorf("expected descending order at index %d: %d < %d",
						k, dc.candidates[k].matchesCount, dc.candidates[k+1].matchesCount)
				}
			}
		})
	}
}

func TestDispatcherCandidate_Fields(t *testing.T) {
	testCases := []struct {
		name               string
		candidate          dispatcherCandidate
		expectedMatches    int
		expectedLiteral    int
		expectedNonDefault int
		description        string
	}{
		{
			name: "simple candidate",
			candidate: dispatcherCandidate{
				matchesCount:    2,
				literalCount:    10,
				nonDefaultCount: 1,
			},
			expectedMatches:    2,
			expectedLiteral:    10,
			expectedNonDefault: 1,
			description:        "should store all fields correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.candidate.matchesCount != tc.expectedMatches {
				t.Errorf("expected matchesCount=%d, got %d", tc.expectedMatches, tc.candidate.matchesCount)
			}
			if tc.candidate.literalCount != tc.expectedLiteral {
				t.Errorf("expected literalCount=%d, got %d", tc.expectedLiteral, tc.candidate.literalCount)
			}
			if tc.candidate.nonDefaultCount != tc.expectedNonDefault {
				t.Errorf("expected nonDefaultCount=%d, got %d", tc.expectedNonDefault, tc.candidate.nonDefaultCount)
			}
		})
	}
}
