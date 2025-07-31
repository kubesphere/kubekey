/*
Copyright 2020 The KubeSphere Authors.

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

package query

import (
	"slices"
	"strconv"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/labels"
)

// Common query parameter names used in API requests
const (
	ParameterName          = "name"          // Name parameter for filtering by resource name
	ParameterLabelSelector = "labelSelector" // Label selector parameter for filtering by labels
	ParameterFieldSelector = "fieldSelector" // Field selector parameter for filtering by fields
	ParameterPage          = "page"          // Page number parameter for pagination
	ParameterLimit         = "limit"         // Items per page parameter for pagination
	ParameterOrderBy       = "sortBy"        // Field to sort results by
	ParameterAscending     = "ascending"     // Sort direction parameter (true for ascending, false for descending)
)

// Query represents API search terms and filtering options
type Query struct {
	Pagination *Pagination // Pagination settings for the query results

	// SortBy specifies which field to sort results by, defaults to FieldCreationTimeStamp
	SortBy string

	// Ascending determines sort direction, defaults to descending (false)
	Ascending bool

	// Filters contains field-value pairs for filtering results
	Filters map[string]string

	// LabelSelector contains the label selector string for filtering by labels
	LabelSelector string
}

// Pagination represents pagination settings for query results
type Pagination struct {
	Limit  int // Number of items per page
	Offset int // Starting offset for the current page
}

// NoPagination represents a query without pagination
var NoPagination = newPagination(-1, 0)

// newPagination creates a new Pagination instance with the given limit and offset
func newPagination(limit int, offset int) *Pagination {
	return &Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

// Selector returns a labels.Selector for the query's label selector
func (q *Query) Selector() labels.Selector {
	selector, err := labels.Parse(q.LabelSelector)
	if err != nil {
		return labels.Everything()
	}

	return selector
}

// AppendLabelSelector adds additional label selectors to the existing query
func (q *Query) AppendLabelSelector(ls map[string]string) error {
	labelsMap, err := labels.ConvertSelectorToLabelsMap(q.LabelSelector)
	if err != nil {
		return err
	}
	q.LabelSelector = labels.Merge(labelsMap, ls).String()
	return nil
}

// GetValidPagination calculates valid start and end indices for pagination
func (p *Pagination) GetValidPagination(total int) (startIndex, endIndex int) {
	// Return all items if no pagination is specified
	if p.Limit == NoPagination.Limit {
		return 0, total
	}

	// Return empty range if pagination parameters are invalid
	if p.Limit < 0 || p.Offset < 0 || p.Offset > total {
		return 0, 0
	}

	startIndex = p.Offset
	endIndex = startIndex + p.Limit

	// Adjust end index if it exceeds total items
	if endIndex > total {
		endIndex = total
	}

	return startIndex, endIndex
}

// New creates a new Query instance with default values
func New() *Query {
	return &Query{
		Pagination: NoPagination,
		SortBy:     "",
		Ascending:  false,
		Filters:    map[string]string{},
	}
}

// Filter represents a single field-value filter pair
type Filter struct {
	Field string `json:"field"` // Field to filter on
	Value string `json:"value"` // Value to filter by
}

// ParseQueryParameter parses query parameters from a RESTful request into a Query struct
func ParseQueryParameter(request *restful.Request) *Query {
	query := New()

	// Parse pagination parameters
	limit, err := strconv.Atoi(request.QueryParameter(ParameterLimit))
	if err != nil {
		limit = -1 // Use default value if undefined
	}
	page, err := strconv.Atoi(request.QueryParameter(ParameterPage))
	if err != nil {
		page = 1 // Use default value if undefined
	}

	query.Pagination = newPagination(limit, (page-1)*limit)

	// Parse sorting parameters
	query.SortBy = DefaultString(request.QueryParameter(ParameterOrderBy), FieldCreationTimeStamp)

	ascending, err := strconv.ParseBool(DefaultString(request.QueryParameter(ParameterAscending), "false"))
	if err != nil {
		query.Ascending = false
	} else {
		query.Ascending = ascending
	}

	// Parse label selector
	query.LabelSelector = request.QueryParameter(ParameterLabelSelector)

	// Parse additional filters
	for key, values := range request.Request.URL.Query() {
		if !slices.Contains([]string{ParameterPage, ParameterLimit, ParameterOrderBy, ParameterAscending, ParameterLabelSelector}, key) {
			value := ""
			if len(values) > 0 {
				value = values[0]
			}
			query.Filters[key] = value
		}
	}

	return query
}

// defaultString returns the default value if the input string is empty
func DefaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
