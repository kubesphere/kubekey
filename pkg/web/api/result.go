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

import (
	"encoding/json"
	"os"
)

const (
	// SchemaLabelSubfix is the label key used to indicate which schema a playbook belongs to.
	SchemaLabelSubfix = "kubekey.kubesphere.io/schema"

	// SchemaProductFile is the predefined file name for storing product information.
	SchemaProductFile = "product.json"
	// SchemaConfigFile is the predefined file name for caching configuration.
	SchemaConfigFile = "config.json"
)

const (
	// ResultSucceed indicates a successful operation result.
	ResultSucceed = "success"
	// ResultFailed indicates a failed operation result.
	ResultFailed = "failed"
	// ResultPending indicates a pending operation result.
	ResultPending = "pending"
)

const (
	// CoreAPIPath defines the base path for core API endpoints in the KubeKey API server.
	// All core resource management routes (inventories, playbooks, etc.) are prefixed with this path.
	CoreAPIPath = "/kapis/"

	// SwaggerAPIPath defines the base path for serving the Swagger UI (OpenAPI documentation).
	// This is used to provide interactive API documentation for the KubeKey API server.
	SwaggerAPIPath = "/swagger-ui/"

	// ResourcesAPIPath defines the base path for resource-related endpoints.
	// This path is used as the prefix for routes that serve static resources, schemas, and related files.
	ResourcesAPIPath = "/resources/"

	// KubeKeyTag is the tag used for KubeKey related resources
	// This tag is used to identify and categorize KubeKey-specific resources
	// in the system, making it easier to filter and manage them
	KubeKeyTag = "kubekey"
	// OpenAPITag is the tag used for OpenAPI documentation
	// This tag helps organize and identify OpenAPI/Swagger documentation
	// related to the KubeKey API endpoints
	OpenAPITag = "api"
	// ResourceTag is the tag used for resource-related endpoints
	// This tag helps organize and identify API endpoints that deal with
	// resource management and operations
	ResourceTag = "resources"

	// StatusOK represents a successful operation status
	// Used to indicate that an API operation completed successfully
	// without any errors or issues
	StatusOK = "ok"
)

// SUCCESS is a global variable representing a successful operation result with a default success message.
// It can be used as a standard response for successful API calls.
var SUCCESS = Result{Message: ResultSucceed}

// Result represents a basic API response structure containing a message field.
// The Message field is typically used to convey error or success information.
type Result struct {
	Message string `description:"error message" json:"message"` // Message provides details about the result or error.
	Result  any    `json:"result"`
}

// SetResult sets the Result field of the Result struct and returns the updated Result.
// This is useful for chaining or for returning a Result with additional data.
func (r Result) SetResult(result any) Result {
	r.Result = result
	return r
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
	Name                 string                `json:"name"`                 // Hostname of the inventory host
	Status               string                `json:"status"`               // Current status of the host
	InternalIPV4         string                `json:"internalIPV4"`         // IPv4 address of the host
	InternalIPV6         string                `json:"internalIPV6"`         // IPv6 address of the host
	SSHHost              string                `json:"sshHost"`              // SSH hostname for connection
	SSHPort              string                `json:"sshPort"`              // SSH port for connection
	SSHUser              string                `json:"sshUser"`              // SSH username for authentication
	SSHPassword          string                `json:"sshPassword"`          // SSH password for authentication
	SSHPrivateKeyContent string                `json:"sshPrivateKeyContent"` // SSH private key content for authentication
	Vars                 map[string]any        `json:"vars"`                 // Additional host variables
	Groups               []InventoryHostGroups `json:"groups"`               // Groups the host belongs to
	Arch                 string                `json:"arch"`                 // Architecture of the host
}

// InventoryHostGroups represents the group information for a host in the inventory.
// Role is the name of the group, and Index is the index of the group to which the host belongs.
type InventoryHostGroups struct {
	Role  string `json:"role"`  // the groups name
	Index int    `json:"index"` // the index of groups which hosts belong to
}

// SchemaFileDataSchema represents the metadata section of a schema file as used in the API layer.
// It contains the main data schema metadata such as title, description, version, namespace, logo, and priority.
type SchemaFileDataSchema struct {
	Title       string `json:"title"`       // Title of the schema
	Description string `json:"description"` // Description of the schema
	Version     string `json:"version"`     // Version of the schema
	Namespace   string `json:"namespace"`   // Namespace of the schema
	Logo        string `json:"logo"`        // Logo URL or identifier
	Priority    int    `json:"priority"`    // Priority for display or ordering
}

// SchemaFile represents the structure of a schema file as used in the API layer.
// It contains the main data schema metadata and a mapping of playbook labels to their paths.
type SchemaFile struct {
	DataSchema   SchemaFileDataSchema `json:"dataSchema"`   // Metadata of the schema file
	PlaybookPath map[string]string    `json:"playbookPath"` // Mapping of playbook labels to their file paths
}

// SchemaTablePlaybook represents the details of a playbook associated with a schema in the response table.
// It includes the path, name, namespace, phase, and result of the playbook.
type SchemaTablePlaybook struct {
	Path      string `json:"path"`      // Path of playbook template.
	Name      string `json:"name"`      // Name of the playbook
	Namespace string `json:"namespace"` // Namespace of the playbook
	Phase     string `json:"phase"`     // Phase of the playbook
	Result    any    `json:"result"`    // Result of the playbook
}

// SchemaTable represents the response table constructed from a schema file.
// It includes metadata fields such as name, title, description, version, namespace, logo, and priority.
// The Playbook field is a map of playbook labels to SchemaTablePlaybook, each representing a playbook reference.
type SchemaTable struct {
	Name        string                         `json:"name"`        // Name of schema, defined by filename
	Title       string                         `json:"title"`       // Title of the schema
	Description string                         `json:"description"` // Description of the schema
	Version     string                         `json:"version"`     // Version of the schema
	Namespace   string                         `json:"namespace"`   // Namespace of the schema
	Logo        string                         `json:"logo"`        // Logo URL or identifier
	Priority    int                            `json:"priority"`    // Priority for display or ordering
	Playbook    map[string]SchemaTablePlaybook `json:"playbook"`    // Map of playbook labels to playbook details
}

// IPTable represents an IP address entry and its SSH status information.
// It indicates whether the IP is a localhost, if SSH is reachable, and if SSH authorization is present.
type IPTable struct {
	IP            string `json:"ip"`            // IP address
	SSHPort       string `json:"sshPort"`       // SSH port
	Localhost     bool   `json:"localhost"`     // Whether the IP is a localhost IP
	SSHReachable  bool   `json:"sshReachable"`  // Whether SSH port is reachable on this IP
	SSHAuthorized bool   `json:"sshAuthorized"` // Whether SSH is authorized for this IP
}

// SchemaFile2Table converts a SchemaFile and its filename into a SchemaTable structure.
// It initializes the SchemaTable fields from the SchemaFile's DataSchema and sets up the Playbook map
// with playbook labels and their corresponding paths. Other playbook fields are left empty for later population.
func SchemaFile2Table(schemaFile SchemaFile, configFile string, filename string) SchemaTable {
	table := SchemaTable{
		Name:        filename,
		Title:       schemaFile.DataSchema.Title,
		Description: schemaFile.DataSchema.Description,
		Version:     schemaFile.DataSchema.Version,
		Namespace:   schemaFile.DataSchema.Namespace,
		Logo:        schemaFile.DataSchema.Logo,
		Priority:    schemaFile.DataSchema.Priority,
		Playbook:    make(map[string]SchemaTablePlaybook),
	}
	// Populate the Playbook map with playbook labels and their paths from the schema file.
	for k, v := range schemaFile.PlaybookPath {
		table.Playbook[k] = SchemaTablePlaybook{
			Path: v,
		}
	}

	// If the schema is for kubernetes.json, try to supplement the version field from the config file.
	if filename == "kubernetes.json" {
		// Attempt to read the config file.
		if configFileBytes, err := os.ReadFile(configFile); err == nil {
			var config map[string]map[string]any
			// Attempt to unmarshal the config file into a map.
			if err := json.Unmarshal(configFileBytes, &config); err == nil {
				// Check if the config contains an entry for "kubernetes.json".
				if v, ok := config[filename]; ok {
					// Check if the "kubernetes" key exists and is a map.
					if kube, ok := v["kubernetes"].(map[string]any); ok {
						// If "kube_version" exists and is a string, set it as the table's version.
						if kubeVer, ok := kube["kube_version"].(string); ok {
							table.Version = kubeVer
						}
					}
				}
			}
		}
	}

	return table
}
