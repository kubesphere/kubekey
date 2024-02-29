package clusterinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClientSet struct {
	Client *kubernetes.Clientset
}

type DumpResources interface {
	GetNodes() []corev1.Node
	GetNamespaces() []corev1.Namespace
	GetConfigMap(namespace string) []corev1.ConfigMap
	GetServices(namespace string) []corev1.Service
	GetSecrets(namespace string) []corev1.Secret
	GetDeployment(namespace string) []appsv1.Deployment
	GetStatefulSets(namespace string) []appsv1.StatefulSet
	GetPods(namespace string) []corev1.Pod
	GetPersistentVolumeClaims(namespace string) []corev1.PersistentVolumeClaim
	GetPersistentVolumes() []corev1.PersistentVolume
	GetStorageClasses() []interface{}
	GetJobs(namespace string) []batchv1.Job
	GetCronJobs(namespace string) []batchv1.CronJob
	GetCRD() []interface{}
	GetClusterResources(namespace string) map[string]map[string]map[string][]interface{}
	GetMultiCluster() ([]v1alpha2.MultiCluster, error)
}

func NewDumpOption(client *kubernetes.Clientset) *ClientSet {
	return &ClientSet{
		Client: client,
	}
}

func (c *ClientSet) GetClusterResources(namespace string) map[string]map[string]map[string][]interface{} {

	clusterResources := map[string]map[string][]interface{}{}
	clusterResources["nodes"] = resourcesClassification(c.GetNodes())
	clusterResources["namespaces"] = resourcesClassification(c.GetNamespaces())
	clusterResources["persistentvolumes"] = resourcesClassification(c.GetPersistentVolumes())
	clusterResources["storageclasses"] = resourcesClassification(c.GetStorageClasses())
	namespaceResources := map[string]map[string][]interface{}{}
	namespaceResources["crds"] = resourcesClassification(c.GetCRD())
	namespaceResources["deployments"] = resourcesClassification(c.GetDeployment(namespace))
	namespaceResources["statefulsets"] = resourcesClassification(c.GetStatefulSets(namespace))
	namespaceResources["pods"] = resourcesClassification(c.GetPods(namespace))
	namespaceResources["services"] = resourcesClassification(c.GetServices(namespace))
	namespaceResources["configmaps"] = resourcesClassification(c.GetConfigMap(namespace))
	namespaceResources["secrets"] = resourcesClassification(c.GetSecrets(namespace))
	namespaceResources["persistentvolumeclaims"] = resourcesClassification(c.GetPersistentVolumeClaims(namespace))
	namespaceResources["jobs"] = resourcesClassification(c.GetJobs(namespace))
	namespaceResources["cronjobs"] = resourcesClassification(c.GetCronJobs(namespace))

	return map[string]map[string]map[string][]interface{}{"clusterResources": clusterResources, "namespaceResources": namespaceResources}

}

func (c *ClientSet) GetCRD() []interface{} {
	raw, err := c.Client.CoreV1().RESTClient().Get().AbsPath("/apis/installer.kubesphere.io/v1alpha1/clusterconfigurations").DoRaw(context.Background())
	if err != nil {
		return nil
	}
	var crd map[string]interface{}
	err = json.Unmarshal(raw, &crd)
	if err != nil {
		return nil
	}
	return crd["items"].([]interface{})
}

func (c *ClientSet) GetNodes() []corev1.Node {
	list, err := c.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}

	return list.Items
}

func (c *ClientSet) GetNamespaces() []corev1.Namespace {

	list, err := c.Client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetConfigMap(namespace string) []corev1.ConfigMap {
	list, err := c.Client.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetMultiCluster() ([]v1alpha2.MultiCluster, error) {
	raw, err := c.Client.CoreV1().RESTClient().Get().AbsPath("/apis/cluster.kubesphere.io/v1alpha1/clusters").DoRaw(context.TODO())
	if err != nil {
		fmt.Println(err, "failed to get cluster config")
		return nil, err
	}
	var cluster v1alpha2.MultiClusterList
	err = json.Unmarshal(raw, &cluster)
	if err != nil {
		fmt.Println(err, "failed to unmarshal cluster config")
		return nil, err
	}

	return cluster.Items, nil
}
func (c *ClientSet) GetDeployment(namespace string) []appsv1.Deployment {

	list, err := c.Client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}

	return list.Items
}

func (c *ClientSet) GetStatefulSets(namespace string) []appsv1.StatefulSet {

	list, err := c.Client.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetPods(namespace string) []corev1.Pod {
	list, err := c.Client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetServices(namespace string) []corev1.Service {
	list, err := c.Client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetSecrets(namespace string) []corev1.Secret {
	list, err := c.Client.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetPersistentVolumeClaims(namespace string) []corev1.PersistentVolumeClaim {

	list, err := c.Client.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}

	return list.Items
}

func (c *ClientSet) GetPersistentVolumes() []corev1.PersistentVolume {
	list, err := c.Client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetStorageClasses() []interface{} {

	raw, err := c.Client.CoreV1().RESTClient().Get().AbsPath("/apis/storage.k8s.io/v1/storageclasses").DoRaw(context.Background())
	if err != nil {
		return nil
	}
	var storageClass map[string]interface{}
	err = json.Unmarshal(raw, &storageClass)
	if err != nil {
		return nil
	}
	return storageClass["items"].([]interface{})
}

func (c *ClientSet) GetJobs(namespace string) []batchv1.Job {
	list, err := c.Client.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}

func (c *ClientSet) GetCronJobs(namespace string) []batchv1.CronJob {
	list, err := c.Client.BatchV1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil
	}
	return list.Items
}
