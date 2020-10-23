/*
Copyright 2020 The KubeSphere Authors.

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

package kubekey

import (
	"context"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/addons/manifests"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

var (
	clusterKubeSphere = template.Must(template.New("cluster.kubesphere.io").Parse(
		dedent.Dedent(`apiVersion: cluster.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  finalizers:
  - finalizer.cluster.kubesphere.io
  labels:
    cluster-role.kubesphere.io/member: ""
    kubesphere.io/managed: "true"
    kubekey.kubesphere.io/name: {{ .Name }}
  name: {{ .Name }} 
spec:
  connection:
    kubeconfig: {{ .Kubeconfig }}
    type: direct
  enable: {{ .Enable }}
  joinFederation: {{ .JoinFedration }}
  provider: kubesphere
    `)))
)

func generateClusterKubeSphere(name string, kubeconfig string, enable, joinFedration bool) (string, error) {
	return util.Render(clusterKubeSphere, util.Data{
		"Name":          name,
		"Kubeconfig":    kubeconfig,
		"Enable":        enable,
		"JoinFedration": joinFedration,
	})
}
func checkClusterRole() (bool, *rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return false, nil, err
	}
	// creates the clientset
	clientset, err := kube.NewForConfig(config)
	if err != nil {
		return false, nil, err
	}
	var hostClusterFlag bool
	if err := clientset.RESTClient().Get().
		AbsPath("/apis/cluster.kubesphere.io/v1alpha1/clusters").
		Name("host").
		Do(context.TODO()).Error(); err == nil {
		hostClusterFlag = true
	}
	return hostClusterFlag, config, nil
}
func newKubeSphereCluster(r *ClusterReconciler, c *kubekeyapiv1alpha1.Cluster) error {
	if hostClusterFlag, config, err := checkClusterRole(); err != nil {
		return err
	} else if hostClusterFlag {
		obj, err := generateClusterKubeSphere(c.Name, "Cg==", false, false)
		if err != nil {
			return err
		}
		if err := manifests.DoServerSideApply(context.TODO(), config, []byte(obj)); err != nil {
			_ = r.Delete(context.TODO(), c)
			return err
		}

		kscluster, err1 := manifests.GetCluster(c.Name)
		if err1 != nil {
			return err
		}
		ownerReferencePatch := fmt.Sprintf(`{"metadata": {"ownerReferences": [{"apiVersion": "%s", "kind": "%s", "name": "%s", "uid": "%s"}]}}`, kscluster.GetAPIVersion(), kscluster.GetKind(), kscluster.GetName(), kscluster.GetUID())
		if err := r.Patch(context.TODO(), c, client.RawPatch(types.MergePatchType, []byte(ownerReferencePatch))); err != nil {
			return err
		}
	}
	return nil
}
func UpdateKubeSphereCluster(mgr *manager.Manager) error {
	if hostClusterFlag, config, err := checkClusterRole(); err != nil {
		return err
	} else if hostClusterFlag {
		obj, err := generateClusterKubeSphere(mgr.ObjName, mgr.Kubeconfig, true, true)
		if err != nil {
			return err
		}
		if err := manifests.DoServerSideApply(context.TODO(), config, []byte(obj)); err != nil {
			return err
		}
	}
	return nil
}

func KubekeyClient() (*kubekeyclientset.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset := kubekeyclientset.NewForConfigOrDie(config)

	return clientset, nil
}

func getCluster(name string) (*kubekeyapiv1alpha1.Cluster, error) {
	clientset, err := KubekeyClient()
	if err != nil {
		return nil, err
	}
	clusterObj, err := clientset.KubekeyV1alpha1().Clusters().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return clusterObj, nil
}

func UpdateClusterConditions(mgr *manager.Manager, step string, startTime, endTime metav1.Time, status bool, index int) error {
	condition := kubekeyapiv1alpha1.Condition{
		Step:      step,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    status,
	}
	if len(mgr.Conditions) < index {
		mgr.Conditions = append(mgr.Conditions, condition)
	} else if len(mgr.Conditions) == index {
		mgr.Conditions[index-1] = condition
	}

	cluster, err := getCluster(mgr.ObjName)
	if err != nil {
		return err
	}

	cluster.Status.Conditions = mgr.Conditions

	if _, err := mgr.ClientSet.KubekeyV1alpha1().Clusters().UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func UpdateStatus(mgr *manager.Manager) error {
	cluster, err := getCluster(mgr.ObjName)
	if err != nil {
		return err
	}
	cluster.Status.Version = mgr.Cluster.Kubernetes.Version
	cluster.Status.NodesCount = len(mgr.AllNodes)
	cluster.Status.MasterCount = len(mgr.MasterNodes)
	cluster.Status.WorkerCount = len(mgr.WorkerNodes)
	cluster.Status.EtcdCount = len(mgr.EtcdNodes)
	cluster.Status.NetworkPlugin = mgr.Cluster.Network.Plugin
	cluster.Status.Nodes = []kubekeyapiv1alpha1.NodeStatus{}

	for _, node := range mgr.AllNodes {
		cluster.Status.Nodes = append(cluster.Status.Nodes, kubekeyapiv1alpha1.NodeStatus{
			InternalIP: node.InternalAddress,
			Hostname:   node.Name,
			Roles:      map[string]bool{"etcd": node.IsEtcd, "master": node.IsMaster, "worker": node.IsWorker},
		})
	}

	if _, err := mgr.ClientSet.KubekeyV1alpha1().Clusters().UpdateStatus(context.TODO(), cluster, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
