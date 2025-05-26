/*
Copyright 2019 The KubeSphere Authors.

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

package api

// SUCCESS represents a successful operation result with a default success message
var SUCCESS = Result{Message: "success"}

// Result represents a basic API response with a message field
type Result struct {
	Message string `description:"error message" json:"message"`
}

// ListResult represents a paginated list response containing items and total count
type ListResult[T any] struct {
	Items      []T `json:"items"`
	TotalItems int `json:"totalItems"`
}

// InventoryHostTable represents a host entry in an inventory with its configuration details
// It includes network information, SSH credentials, and group membership
type InventoryHostTable struct {
	Name          string         `json:"name"`          // Hostname of the inventory host
	Status        string         `json:"status"`        // Current status of the host
	InternalIPV4  string         `json:"internalIPV4"`  // IPv4 address of the host
	InternalIPV6  string         `json:"internalIPV6"`  // IPv6 address of the host
	SSHHost       string         `json:"sshHost"`       // SSH hostname for connection
	SSHPort       string         `json:"sshPort"`       // SSH port for connection
	SSHUser       string         `json:"sshUser"`       // SSH username for authentication
	SSHPassword   string         `json:"sshPassword"`   // SSH password for authentication
	SSHPrivateKey string         `json:"sshPrivateKey"` // SSH private key for authentication
	Vars          map[string]any `json:"vars"`          // Additional host variables
	Groups        []string       `json:"groups"`        // Groups the host belongs to
	Arch          string         `json:"arch"`          // Architecture of the host
}
