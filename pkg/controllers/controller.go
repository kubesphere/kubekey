package controllers

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Controllers store all controllers
var Controllers []controller

// Register controller
func Register(reconciler controller) error {
	for _, c := range Controllers {
		if c.Name() == reconciler.Name() {
			return fmt.Errorf("%s has register", reconciler.Name())
		}
	}
	Controllers = append(Controllers, reconciler)

	return nil
}

// controller should add in ctrl.manager
type controller interface {
	reconcile.Reconciler
	// setup reconcile with manager
	SetupWithManager(mgr ctrl.Manager, o ctrlcontroller.Options) error
	// the Name of controller
	Name() string
}
