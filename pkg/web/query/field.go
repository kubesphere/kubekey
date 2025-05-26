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

// Field represents a query field name used for filtering and sorting
type Field string

// Value represents a query field value used for filtering
type Value string

const (
	// FieldName represents the name field of a resource
	FieldName = "name"
	// FieldNames represents multiple names field of resources
	FieldNames = "names"
	// FieldUID represents the unique identifier field of a resource
	FieldUID = "uid"
	// FieldCreationTimeStamp represents the creation timestamp field of a resource
	FieldCreationTimeStamp = "creationTimestamp"
	// FieldCreateTime represents the creation time field of a resource
	FieldCreateTime = "createTime"
	// FieldLastUpdateTimestamp represents the last update timestamp field of a resource
	FieldLastUpdateTimestamp = "lastUpdateTimestamp"
	// FieldUpdateTime represents the update time field of a resource
	FieldUpdateTime = "updateTime"
	// FieldLabel represents the label field of a resource
	FieldLabel = "label"
	// FieldAnnotation represents the annotation field of a resource
	FieldAnnotation = "annotation"
	// FieldNamespace represents the namespace field of a resource
	FieldNamespace = "namespace"
	// FieldStatus represents the status field of a resource
	FieldStatus = "status"
	// FieldOwnerReference represents the owner reference field of a resource
	FieldOwnerReference = "ownerReference"
	// FieldOwnerKind represents the owner kind field of a resource
	FieldOwnerKind = "ownerKind"
)
