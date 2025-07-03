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

// SUCCESS is a global variable representing a successful operation result with a default success message.
// It can be used as a standard response for successful API calls.
var SUCCESS = Result{Message: "success"}

// Result represents a basic API response structure containing a message field.
// The Message field is typically used to convey error or success information.
type Result struct {
	Message string `description:"error message" json:"message"` // Message provides details about the result or error.
}

// ListResult is a generic struct representing a paginated list response.
// T is a type parameter for the type of items in the list.
// Items contains the list of results, and TotalItems indicates the total number of items available.
type ListResult[T any] struct {
	Items      []T `json:"items"`      // List of items of type T
	TotalItems int `json:"totalItems"` // Total number of items available
}

// InventoryHostTable represents a host entry in an inventory with its configuration details.
// It includes network information, SSH credentials, group membership, and architecture.
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

// SchemaTable represents schema metadata for a resource.
// It includes fields such as name, type, title, description, version, namespace, logo, priority, and associated playbooks.
// The Playbook field is a slice of SchemaTablePlaybook, each representing a playbook reference.
type SchemaTable struct {
	Name        string                `json:"name"`        // Name of schema, defined by filename
	SchemaType  string                `json:"schemaType"`  // Type of the schema (e.g., CRD, built-in)
	Title       string                `json:"title"`       // Title of the schema
	Description string                `json:"description"` // Description of the schema
	Version     string                `json:"version"`     // Version of the schema
	Namespace   string                `json:"namespace"`   // Namespace of the schema
	Logo        string                `json:"logo"`        // Logo URL or identifier
	Priority    int                   `json:"priority"`    // Priority for display or ordering
	Playbook    []SchemaTablePlaybook `json:"playbook"`    // List of reference playbooks
}

// SchemaTablePlaybook represents a reference to a playbook associated with a schema.
// It includes the playbook's name, namespace, and phase.
type SchemaTablePlaybook struct {
	Name      string `json:"name"`      // Name of the playbook
	Namespace string `json:"namespace"` // Namespace of the playbook
	Phase     string `json:"phase"`     // Phase of the playbook
}

// IPTable represents an IP address entry and its SSH status information.
// It indicates whether the IP is a localhost, if SSH is reachable, and if SSH authorization is present.
type IPTable struct {
	IP            string `json:"ip"`            // IP address
	Localhost     bool   `json:"localhost"`     // Whether the IP is a localhost IP
	SSHReachable  bool   `json:"sshReachable"`  // Whether SSH port is reachable on this IP
	SSHAuthorized bool   `json:"sshAuthorized"` // Whether SSH is authorized for this IP
}
