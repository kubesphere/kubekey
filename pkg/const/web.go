package _const

import "k8s.io/apimachinery/pkg/runtime"

const (
	// APIPath defines the base path for API endpoints in the KubeKey API server.
	// This path is used as the prefix for all API routes, including those for
	// managing inventories, playbooks, and other KubeKey resources.
	APIPath = "/kapis/"

	// KubeKeyTag is the tag used for KubeKey related resources
	// This tag is used to identify and categorize KubeKey-specific resources
	// in the system, making it easier to filter and manage them
	KubeKeyTag = "kubekey"

	// OpenAPITag is the tag used for OpenAPI documentation
	// This tag helps organize and identify OpenAPI/Swagger documentation
	// related to the KubeKey API endpoints
	OpenAPITag = "api"

	// StatusOK represents a successful operation status
	// Used to indicate that an API operation completed successfully
	// without any errors or issues
	StatusOK = "ok"
)

// SUCCESS is a predefined successful result
// This is a convenience variable that provides a standard success response
// for API operations that don't need to return specific data
var SUCCESS = Result{Message: "success"}

// Result represents a generic API response with a message
// This type is used for simple API responses that only need to convey
// a status message, such as success or error notifications
type Result struct {
	Message string `description:"error message" json:"message"`
}

// ListResult represents a paginated list response containing items and total count
// This type is used for API responses that return a list of items with pagination
// support, allowing clients to handle large datasets efficiently
type ListResult struct {
	Items      []runtime.Object `json:"items"`      // The list of items in the current page
	TotalItems int              `json:"totalItems"` // The total number of items across all pages
}
