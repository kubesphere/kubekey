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

package controllers

import ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"

type Options struct {
	ControllerGates []string

	ctrlcontroller.Options
}

// IsControllerEnabled check if a specified controller enabled or not.
func (o Options) IsControllerEnabled(name string) bool {
	hasStar := false
	for _, ctrl := range o.ControllerGates {
		if ctrl == name {
			return true
		}
		if ctrl == "-"+name {
			return false
		}
		if ctrl == "*" {
			hasStar = true
		}
	}

	return hasStar
}
