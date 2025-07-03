package _const

import "os"

// Environment represents an environment variable with its name and default value
type Environment struct {
	env string // environment variable name
	def string // default value if environment variable is not set
}

var (
	// Shell specifies which shell operator uses in local connector
	Shell = Environment{env: "SHELL", def: "/bin/bash"}

	// ExecutorVerbose specifies the verbosity level used in playbook pod
	ExecutorVerbose = Environment{env: "EXECUTOR_VERBOSE"}
	// ExecutorImage specifies the container image used in playbook pod
	ExecutorImage = Environment{env: "EXECUTOR_IMAGE", def: "docker.io/kubesphere/executor:latest"}
	// ExecutorImagePullPolicy specifies the image pull policy used in playbook pod
	ExecutorImagePullPolicy = Environment{env: "EXECUTOR_IMAGE_PULLPOLICY"}
	// ExecutorClusterRole specifies the cluster role used in playbook pod
	ExecutorClusterRole = Environment{env: "EXECUTOR_CLUSTERROLE"}

	// CapkkGroupControlPlane specifies the control plane groups for capkk playbook
	CapkkGroupControlPlane = Environment{env: "CAPKK_GROUP_CONTROLPLANE", def: "kube_control_plane"}
	// CapkkGroupWorker specifies the worker groups for capkk playbook
	CapkkGroupWorker = Environment{env: "CAPKK_GROUP_WORKER", def: "kube_worker"}
	// CapkkVolumeBinary specifies a persistent volume containing the CAPKKBinarydir for capkk playbook, used in offline installer
	CapkkVolumeBinary = Environment{env: "CAPKK_VOLUME_BINARY"}
	// CapkkVolumeProject specifies a persistent volume containing the CAPKKProjectdir for capkk playbook
	CapkkVolumeProject = Environment{env: "CAPKK_VOLUME_PROJECT"}
	// CapkkVolumeWorkdir specifies the working directory for capkk playbook
	CapkkVolumeWorkdir = Environment{env: "CAPKK_VOLUME_WORKDIR"}
)

// Getenv retrieves the value of the environment variable. If the environment variable is not set,
// it returns the default value specified in the Environment struct.
func Getenv(env Environment) string {
	val, ok := os.LookupEnv(env.env)
	if !ok {
		return env.def
	}
	return val
}
