package controller

import (
	"github.com/kubekey/pkg/controller/k2cluster"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, k2cluster.Add)
}
