/*
Copyright 2020 KubeSphere Authors

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
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/web/api"
)

// CompareFunc is a generic function type that compares two objects of type T
// Returns true if left is greater than right
type CompareFunc[T any] func(T, T, Field) bool

// FilterFunc is a generic function type that filters objects of type T
// Returns true if the object matches the filter criteria
type FilterFunc[T any] func(T, Filter) bool

// TransformFunc is a generic function type that transforms objects of type T
// Returns the transformed object
type TransformFunc[T any] func(T) T

// DefaultList processes a list of objects with filtering, sorting, and pagination
// Parameters:
//   - objects: The list of objects to process
//   - q: Query parameters including filters, sorting, and pagination
//   - compareFunc: Function to compare objects for sorting
//   - filterFunc: Function to filter objects
//   - transformFuncs: Optional functions to transform objects
//
// Returns a ListResult containing the processed objects
func DefaultList[T any](objects []T, q *Query, compareFunc CompareFunc[T], filterFunc FilterFunc[T], transformFuncs ...TransformFunc[T]) *api.ListResult[T] {
	// selected matched ones
	filtered := make([]T, 0)
	for _, object := range objects {
		selected := true
		for field, value := range q.Filters {
			if !filterFunc(object, Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			for _, transform := range transformFuncs {
				object = transform(object)
			}
			filtered = append(filtered, object)
		}
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !q.Ascending {
			return compareFunc(filtered[i], filtered[j], q.SortBy)
		}
		return !compareFunc(filtered[i], filtered[j], q.SortBy)
	})

	total := len(filtered)

	if q.Pagination == nil {
		q.Pagination = NoPagination
	}

	start, end := q.Pagination.GetValidPagination(total)

	return &api.ListResult[T]{
		TotalItems: len(filtered),
		Items:      filtered[start:end],
	}
}

// DefaultObjectMetaCompare compares two metav1.Object instances
// Returns true if left is greater than right based on the specified sort field
// Supports sorting by name or creation timestamp
func DefaultObjectMetaCompare(left, right metav1.Object, sortBy Field) bool {
	switch sortBy {
	// ?sortBy=name
	case FieldName:
		return left.GetName() > right.GetName()
	//	?sortBy=creationTimestamp
	default:
		// compare by name if creation timestamp is equal
		if left.GetCreationTimestamp().Time.Equal(right.GetCreationTimestamp().Time) {
			return left.GetName() > right.GetName()
		}
		return left.GetCreationTimestamp().After(right.GetCreationTimestamp().Time)
	}
}

// DefaultObjectMetaFilter filters metav1.Object instances based on various criteria
// Supports filtering by:
//   - Names: Exact match against a comma-separated list of names
//   - Name: Partial match against object name
//   - UID: Exact match against object UID
//   - Namespace: Exact match against namespace
//   - OwnerReference: Match against owner reference UID
//   - OwnerKind: Match against owner reference kind
//   - Annotation: Match against annotations using label selector syntax
//   - Label: Match against labels using label selector syntax
func DefaultObjectMetaFilter(item metav1.Object, filter Filter) bool {
	switch filter.Field {
	case FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if item.GetName() == name {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&name=default
	case FieldName:
		return strings.Contains(item.GetName(), string(filter.Value))
	// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case FieldUID:
		return string(item.GetUID()) == string(filter.Value)
	// /deployments?page=1&limit=10&namespace=kubesphere-system
	case FieldNamespace:
		return item.GetNamespace() == string(filter.Value)
	// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case FieldOwnerReference:
		for _, ownerReference := range item.GetOwnerReferences() {
			if string(ownerReference.UID) == string(filter.Value) {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&ownerKind=Workspace
	case FieldOwnerKind:
		for _, ownerReference := range item.GetOwnerReferences() {
			if ownerReference.Kind == string(filter.Value) {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&annotation=openpitrix_runtime
	case FieldAnnotation:
		return labelMatch(item.GetAnnotations(), string(filter.Value))
	// /namespaces?page=1&limit=10&label=kubesphere.io/workspace:system-workspace
	case FieldLabel:
		return labelMatch(item.GetLabels(), string(filter.Value))
	// not supported filter
	default:
		return true
	}
}

// labelMatch checks if a map of labels/annotations matches a label selector
// Returns true if the map matches the selector, false otherwise
func labelMatch(m map[string]string, filter string) bool {
	labelSelector, err := labels.Parse(filter)
	if err != nil {
		klog.Warningf("invalid labelSelector %s: %s", filter, err)
		return false
	}
	return labelSelector.Matches(labels.Set(m))
}
