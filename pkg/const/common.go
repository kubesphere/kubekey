/*
Copyright 2023 The KubeSphere Authors.

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

package _const

// LocalHostName is the  default local host name in inventory.
const LocalHostName = "localhost"

// variable specific key of top level
const (
	VariableHostName = "inventory_name"
	// VariableGlobalHosts the key is host_name, the value is host_var which defined in inventory.
	VariableGlobalHosts = "inventory_hosts"
	// VariableGroups the key is group's name, the value is a host_name slice
	VariableGroups = "groups"
)
