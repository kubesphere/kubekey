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

import "reflect"

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

// GetFieldByJSONTag returns the value of the struct field whose JSON tag matches the given field name (filed).
// If not found by JSON tag, it tries to find the field by its struct field name.
// The function expects obj to be a struct or a pointer to a struct.
func GetFieldByJSONTag(obj reflect.Value, filed string) reflect.Value {
	// If obj is a pointer, get the element it points to
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	// If obj is not a struct, return zero Value
	if obj.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	typ := obj.Type()
	// Iterate over all struct fields
	for i := range obj.NumField() {
		structField := typ.Field(i)
		jsonTag := structField.Tag.Get("json")
		// The tag may have options, e.g. "name,omitempty"
		// Check for exact match or prefix match before comma
		if jsonTag == filed ||
			(jsonTag != "" && jsonTag == filed+",omitempty") ||
			(jsonTag != "" && len(jsonTag) >= len(filed) &&
				jsonTag[:len(filed)] == filed &&
				(len(jsonTag) == len(filed) || jsonTag[len(filed)] == ',')) {
			// Return the field value if the JSON tag matches
			return obj.Field(i)
		}
	}
	// If not found by json tag, try by field name (case-sensitive)
	if f := obj.FieldByName(filed); f.IsValid() {
		return f
	}
	// Return zero Value if not found
	return reflect.Value{}
}
