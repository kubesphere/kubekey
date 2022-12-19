/*
 Copyright 2022 The KubeSphere Authors.

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

package collections

import (
	"sort"

	"sigs.k8s.io/cluster-api/util/conditions"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
)

// KKInstances is a set of KKInstances.
type KKInstances map[string]*infrav1.KKInstance

// kkInstancesByCreationTimestamp sorts a list of KKInstance by creation timestamp, using their names as a tie breaker.
type kkInstancesByCreationTimestamp []*infrav1.KKInstance

func (o kkInstancesByCreationTimestamp) Len() int      { return len(o) }
func (o kkInstancesByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o kkInstancesByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// New creates an empty KKInstances.
func New() KKInstances {
	return make(KKInstances)
}

// FromKKInstances creates a KKInstance from a list of values.
func FromKKInstances(kkInstances ...*infrav1.KKInstance) KKInstances {
	ss := make(KKInstances, len(kkInstances))
	ss.Insert(kkInstances...)
	return ss
}

// FromKKInstanceList creates a KKInstances from the given KKInstanceList.
func FromKKInstanceList(kkInstanceList *infrav1.KKInstanceList) KKInstances {
	ss := make(KKInstances, len(kkInstanceList.Items))
	for i := range kkInstanceList.Items {
		ss.Insert(&kkInstanceList.Items[i])
	}
	return ss
}

// ToKKInstanceList creates a KKInstanceList from the given KKInstances.
func ToKKInstanceList(kkInstances KKInstances) infrav1.KKInstanceList {
	ml := infrav1.KKInstanceList{}
	for _, m := range kkInstances {
		ml.Items = append(ml.Items, *m)
	}
	return ml
}

// Insert adds items to the set.
func (s KKInstances) Insert(kkInstances ...*infrav1.KKInstance) {
	for i := range kkInstances {
		if kkInstances[i] != nil {
			m := kkInstances[i]
			s[m.Name] = m
		}
	}
}

// Difference returns a copy without KKInstances that are in the given collection.
func (s KKInstances) Difference(kkInstances KKInstances) KKInstances {
	return s.Filter(func(m *infrav1.KKInstance) bool {
		_, found := kkInstances[m.Name]
		return !found
	})
}

// SortedByCreationTimestamp returns the KKInstances sorted by creation timestamp.
func (s KKInstances) SortedByCreationTimestamp() []*infrav1.KKInstance {
	res := make(kkInstancesByCreationTimestamp, 0, len(s))
	for _, value := range s {
		res = append(res, value)
	}
	sort.Sort(res)
	return res
}

// UnsortedList returns the slice with contents in random order.
func (s KKInstances) UnsortedList() []*infrav1.KKInstance {
	res := make([]*infrav1.KKInstance, 0, len(s))
	for _, value := range s {
		res = append(res, value)
	}
	return res
}

// Len returns the size of the set.
func (s KKInstances) Len() int {
	return len(s)
}

// newFilteredKKInstanceCollection creates a KKInstances from a filtered list of values.
func newFilteredKKInstanceCollection(filter Func, kkInstances ...*infrav1.KKInstance) KKInstances {
	ss := make(KKInstances, len(kkInstances))
	for i := range kkInstances {
		m := kkInstances[i]
		if filter(m) {
			ss.Insert(m)
		}
	}
	return ss
}

// Filter returns a KKInstances containing only the KKInstances that match all of the given KKInstanceFilters.
func (s KKInstances) Filter(filters ...Func) KKInstances {
	return newFilteredKKInstanceCollection(And(filters...), s.UnsortedList()...)
}

// AnyFilter returns a KKInstances containing only the KKInstances that match any of the given KKInstanceFilters.
func (s KKInstances) AnyFilter(filters ...Func) KKInstances {
	return newFilteredKKInstanceCollection(Or(filters...), s.UnsortedList()...)
}

// Oldest returns the KKInstances with the oldest CreationTimestamp.
func (s KKInstances) Oldest() *infrav1.KKInstance {
	if len(s) == 0 {
		return nil
	}
	return s.SortedByCreationTimestamp()[0]
}

// Newest returns the KKInstance with the most recent CreationTimestamp.
func (s KKInstances) Newest() *infrav1.KKInstance {
	if len(s) == 0 {
		return nil
	}
	return s.SortedByCreationTimestamp()[len(s)-1]
}

// DeepCopy returns a deep copy.
func (s KKInstances) DeepCopy() KKInstances {
	result := make(KKInstances, len(s))
	for _, m := range s {
		result.Insert(m.DeepCopy())
	}
	return result
}

// ConditionGetters returns the slice with KKInstances converted into conditions.Getter.
func (s KKInstances) ConditionGetters() []conditions.Getter {
	res := make([]conditions.Getter, 0, len(s))
	for _, v := range s {
		value := *v
		res = append(res, &value)
	}
	return res
}

// Names returns a slice of the names of each KKInstance in the collection.
// Useful for logging and test assertions.
func (s KKInstances) Names() []string {
	names := make([]string, 0, s.Len())
	for _, m := range s {
		names = append(names, m.Name)
	}
	return names
}
