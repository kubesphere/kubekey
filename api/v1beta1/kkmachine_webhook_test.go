/*
 Copyright 2022 The KubeSphere Authors.

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

package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utildefaulting "sigs.k8s.io/cluster-api/util/defaulting"
)

func TestKKMachine_Default(t *testing.T) {
	g := NewWithT(t)
	kkm := &KKMachine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foobar",
		},
	}

	t.Run("for KKMachine", utildefaulting.DefaultValidateTest(kkm))
	kkm.Default()

	g.Expect(kkm.Spec.ContainerManager.Type).To(Equal(ContainerdType))
	g.Expect(kkm.Spec.ContainerManager.Version).To(Equal(DefaultContainerdVersion))
	g.Expect(kkm.Spec.ContainerManager.CRICTLVersion).To(Equal(DefaultCrictlVersion))

	kkm2 := &KKMachine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foobar",
		},
		Spec: KKMachineSpec{
			ContainerManager: ContainerManager{
				Type:          ContainerdType,
				Version:       "v1.6.4",
				CRICTLVersion: "1.24.0",
			},
		},
	}

	t.Run("for KKMachine2", utildefaulting.DefaultValidateTest(kkm2))
	kkm2.Default()

	g.Expect(kkm2.Spec.ContainerManager.Type).To(Equal(ContainerdType))
	g.Expect(kkm2.Spec.ContainerManager.Version).To(Equal("1.6.4"))
	g.Expect(kkm2.Spec.ContainerManager.CRICTLVersion).To(Equal("v1.24.0"))
}
