package internal

// Standard output results for module execution.
// These constants are used to standardize the messages returned by modules for their basic result states.
const (
	// StdoutSuccess is the value returned when a module finishes successfully.
	// Usually, this means the module executed its task as expected without errors.
	StdoutSuccess = "success"

	// StdoutFailed is the value returned when a module fails to complete its task.
	// Any fatal error in execution will typically result in this status.
	StdoutFailed = "failed"

	// StdoutSkip is used when a module decides to skip its operation, for example,
	// due to a "when" condition being false, or a particular host/state not requiring action.
	StdoutSkip = "skip"
)

// Standard error messages for module execution failures.
// Each error string describes a specific reason for failure, facilitating debugging and log interpretation.
const (
	// StderrGetConnector indicates that the connector (responsible for host communication)
	// could not be obtained or initialized. This can occur if connection details are missing,
	// invalid, or initialization logic fails.
	StderrGetConnector = "failed to get connector"

	// StderrGetHostVariable indicates a failure to retrieve the required host variables.
	// The module likely depends on host-scoped data that could not be loaded or parsed.
	StderrGetHostVariable = "failed to get host variable"

	// StderrParseArgument means module arguments could not be parsed.
	// For example, if user-provided arguments are malformed, missing, or not valid YAML/JSON/etc.
	StderrParseArgument = "failed to parse argument"

	// StderrUnsupportArgs means the arguments provided are syntactically valid but
	// not recognized or not supported by the current module's implementation.
	StderrUnsupportArgs = "unsupport args"

	// StderrGetPlaybook is returned when the playbook (the sequence of tasks or operations)
	// could not be loaded, parsed, or found. Module execution often depends on an existing playbook context.
	StderrGetPlaybook = "failed to get playbook"
)
