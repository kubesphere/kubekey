package proxy

import "net/http"

type router struct {
	path     string          // The path of the action
	pathExpr *pathExpression // cached compilation of rootPath as RegExp
	handlers map[string]http.HandlerFunc
}

func (r router) matchPath(url string) bool {
	return r.pathExpr.Matcher.MatchString(url)
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
