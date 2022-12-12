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

package collections

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
)

// GetFilteredKKInstancesForKKCluster returns a list of kkInstances that can be filtered or not.
// If no filter is supplied then all kkInstances associated with the target kkCluster are returned.
func GetFilteredKKInstancesForKKCluster(ctx context.Context, c client.Reader, kkCluster *infrav1.KKCluster, filters ...Func) (KKInstances, error) {
	ml := &infrav1.KKInstanceList{}
	if err := c.List(
		ctx,
		ml,
		client.InNamespace(kkCluster.Namespace),
		client.MatchingLabels{
			infrav1.KKClusterLabelName: kkCluster.Name,
		},
	); err != nil {
		return nil, errors.Wrap(err, "failed to list machines")
	}

	kkInstances := FromKKInstanceList(ml)
	return kkInstances.Filter(filters...), nil
}
