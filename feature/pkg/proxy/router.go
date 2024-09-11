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

import "net/http"

type router struct {
	path     string          // The path of the action
	pathExpr *pathExpression // cached compilation of rootPath as RegExp
	handlers map[string]http.HandlerFunc
}

// Types and functions to support the sorting of Dispatchers

type dispatcherCandidate struct {
	router          router
	finalMatch      string
	matchesCount    int // the number of capturing groups
	literalCount    int // the number of literal characters (means those not resulting from template variable substitution)
	nonDefaultCount int // the number of capturing groups with non-default regular expressions (i.e. not ‘([^  /]+?)’)
}
type sortableDispatcherCandidates struct {
	candidates []dispatcherCandidate
}

func (dc *sortableDispatcherCandidates) Len() int {
	return len(dc.candidates)
}

func (dc *sortableDispatcherCandidates) Swap(i, j int) {
	dc.candidates[i], dc.candidates[j] = dc.candidates[j], dc.candidates[i]
}

func (dc *sortableDispatcherCandidates) Less(i, j int) bool {
	ci := dc.candidates[i]
	cj := dc.candidates[j]
	// primary key
	if ci.matchesCount < cj.matchesCount {
		return true
	}
	if ci.matchesCount > cj.matchesCount {
		return false
	}
	// secundary key
	if ci.literalCount < cj.literalCount {
		return true
	}
	if ci.literalCount > cj.literalCount {
		return false
	}
	// tertiary key
	return ci.nonDefaultCount < cj.nonDefaultCount
}
