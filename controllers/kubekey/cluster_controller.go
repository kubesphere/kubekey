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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	yamlV2 "gopkg.in/yaml.v2"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	kubekeyv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	CreateCluster = "create cluster"
	AddNodes      = "add nodes"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=storage.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=*
// +kubebuilder:rbac:groups=installer.kubesphere.io,resources=clusterconfigurations,verbs=*
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=auditing.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=autoscaling,resources=*,verbs=*
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=config.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=core.kubefed.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=devops.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=events.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=batch,resources=*,verbs=*
// +kubebuilder:rbac:groups=extensions,resources=*,verbs=*
// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=jaegertracing.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=logging.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=*,verbs=*
// +kubebuilder:rbac:groups=networking.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=notification.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=policy,resources=*,verbs=*
// +kubebuilder:rbac:groups=storage.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=cluster.kubesphere.io,resources=*,verbs=*

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cluster", req.NamespacedName)
	// Fetch the Cluster object
	cluster := &kubekeyv1alpha1.Cluster{}
	cmFound := &corev1.ConfigMap{}
	jobFound := &batchv1.Job{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if kubeErr.IsNotFound(err) {
			log.Info("Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Cluster")
		return ctrl.Result{}, err
	}

	// init a new cluster
	errFoundCreate := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-create-cluster", cluster.Name), Namespace: "kubekey-system"}, jobFound)
	if errFoundCreate == nil {
		if jobFound.Status.Failed != 0 {
			log.Error(errors.New(fmt.Sprintf("Failed to create new Cluster: %s", cluster.Name)), fmt.Sprintf("Job failed: Job.Namespace %s, Job.Name %s", "kubekey-system", fmt.Sprintf("%s-create-cluster", cluster.Name)))
		}
	}
	if kubeErr.IsNotFound(errFoundCreate) && cluster.Status.NodesCount == 0 {
		// create kubesphere cluster
		if err := newKubeSphereCluster(r, cluster); err != nil {
			return ctrl.Result{}, err
		}

		// Check if the configmap already exists, if not create a new one
		if err := r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: "kubekey-system"}, cmFound); err != nil && kubeErr.IsNotFound(err) {
			// Define a new configmap
			cmCluster := r.configMapForCluster(cluster)
			log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
			if err := r.Create(ctx, cmCluster); err != nil {
				log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
				return ctrl.Result{}, nil
			}
		} else if err != nil {
			log.Error(err, "Failed to get ConfigMap")
			return ctrl.Result{}, err
		} else {
			_ = r.Delete(ctx, cmFound)
			cmCluster := r.configMapForCluster(cluster)
			log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
			if err := r.Create(ctx, cmCluster); err != nil {
				log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
				return ctrl.Result{}, nil
			}
		}

		// Check if the job already exists, if not create a new one
		if err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-create-cluster", cluster.Name), Namespace: "kubekey-system"}, jobFound); err != nil && kubeErr.IsNotFound(err) {
			// Define a new Job
			jobCluster := r.jobForCluster(cluster, CreateCluster)
			log.Info("Creating a new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
			if err := r.Create(ctx, jobCluster); err != nil {
				log.Error(err, "Failed to create new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
				return ctrl.Result{}, nil
			}
			if err := updateStatusRunner(r, cluster, CreateCluster); err != nil {
				return ctrl.Result{}, err
			}
		} else if err != nil {
			log.Error(err, "Failed to get Job")
			return ctrl.Result{}, err
		}
	}

	// add nodes to cluster
	if cluster.Status.NodesCount != 0 && (len(cluster.Spec.Hosts) > cluster.Status.NodesCount) {
		// Check if the job already exists, if not create a new one
		if err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-add-nodes", cluster.Name), Namespace: "kubekey-system"}, jobFound); err != nil && kubeErr.IsNotFound(err) {
			if err := createClusterConfigMap(r, cluster, log); err != nil {
				return ctrl.Result{}, err
			}
			// Define a new Job
			jobCluster := r.jobForCluster(cluster, AddNodes)
			log.Info("Creating a new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
			if err := r.Create(ctx, jobCluster); err != nil {
				log.Error(err, "Failed to create new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
				return ctrl.Result{}, nil
			}
			if err := updateStatusRunner(r, cluster, AddNodes); err != nil {
				return ctrl.Result{}, nil
			}
		} else if err != nil {
			log.Error(err, "Failed to get Job")
			return ctrl.Result{}, err
		} else if jobFound.Status.Succeeded == 1 {
			if err1 := r.Delete(ctx, jobFound); err1 != nil {
				return ctrl.Result{}, nil
			}
			if err := createClusterConfigMap(r, cluster, log); err != nil {
				return ctrl.Result{}, err
			}
			// Define a new Job
			jobCluster := r.jobForCluster(cluster, AddNodes)
			log.Info("Creating a new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
			if err := r.Create(ctx, jobCluster); err != nil {
				log.Error(err, "Failed to create new Job", "Job.Namespace", jobCluster.Namespace, "Job.Name", jobCluster.Name)
				return ctrl.Result{}, nil
			}
			if err := updateStatusRunner(r, cluster, AddNodes); err != nil {
				return ctrl.Result{}, nil
			}
		} else if jobFound.Status.Failed != 0 {
			log.Error(errors.New(fmt.Sprintf("Failed to add nodes to Cluster: %s", cluster.Name)), fmt.Sprintf("Job failed: Job.Namespace %s, Job.Name %s", "kubekey-system", fmt.Sprintf("%s-add-nodes", cluster.Name)))
		}
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubekeyv1alpha1.Cluster{}).
		Complete(r)
}

func (r *ClusterReconciler) configMapForCluster(c *kubekeyv1alpha1.Cluster) *corev1.ConfigMap {
	type Metadata struct {
		Name string `yaml:"name" json:"name,omitempty"`
	}
	clusterConfiguration := struct {
		ApiVersion string                      `yaml:"apiVersion" json:"apiVersion,omitempty"`
		Kind       string                      `yaml:"kind" json:"kind,omitempty"`
		Metadata   Metadata                    `yaml:"metadata" json:"metadata,omitempty"`
		Spec       kubekeyv1alpha1.ClusterSpec `yaml:"spec" json:"spec,omitempty"`
	}{ApiVersion: c.APIVersion, Kind: c.Kind, Metadata: Metadata{Name: c.Name}, Spec: c.Spec}

	clusterStr, _ := yamlV2.Marshal(clusterConfiguration)

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: "kubekey-system",
			Labels:    map[string]string{"kubekey.kubesphere.io/name": c.Name},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: c.APIVersion,
				Kind:       c.Kind,
				Name:       c.Name,
				UID:        c.UID,
			}},
		},
		Data: map[string]string{"cluster.yaml": string(clusterStr)},
	}
	return cm
}

func (r *ClusterReconciler) jobForCluster(c *kubekeyv1alpha1.Cluster, action string) *batchv1.Job {
	var (
		backoffLimit int32 = 0
		name         string
		args         []string
	)
	if action == CreateCluster {
		name = fmt.Sprintf("%s-create-cluster", c.Name)
		args = []string{"create", "cluster", "-f", "/home/kubekey/config/cluster.yaml", "-y", "--in-cluster", "true"}
	} else if action == AddNodes {
		name = fmt.Sprintf("%s-add-nodes", c.Name)
		args = []string{"add", "nodes", "-f", "/home/kubekey/config/cluster.yaml", "-y", "--in-cluster", "true"}
	}

	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace("kubekey-system"),
		client.MatchingLabels{"control-plane": "controller-manager"},
	}
	_ = r.List(context.TODO(), podlist, listOpts...)
	nodeName := podlist.Items[0].Spec.NodeName
	var image string
	for _, container := range podlist.Items[0].Spec.Containers {
		if container.Name == "manager" {
			image = container.Image
		}
	}
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kubekey-system",
			Labels:    map[string]string{"kubekey.kubesphere.io/name": c.Name},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: c.APIVersion,
				Kind:       c.Kind,
				Name:       c.Name,
				UID:        c.UID,
			}},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: c.Name,
								},
								Items: []corev1.KeyToPath{{
									Key:  "cluster.yaml",
									Path: "cluster.yaml",
								}},
							},
						},
					}},
					Containers: []corev1.Container{{
						Name:            "runner",
						Image:           image,
						ImagePullPolicy: "Always",
						Command:         []string{"/home/kubekey/kk"},
						Args:            args,
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config",
							MountPath: "/home/kubekey/config",
						}},
					}},
					NodeName:           nodeName,
					ServiceAccountName: "default",
					RestartPolicy:      "Never",
				},
			},
		},
	}

	return job
}

func updateStatusRunner(r *ClusterReconciler, cluster *kubekeyv1alpha1.Cluster, action string) error {
	var (
		name string
	)
	if action == CreateCluster {
		name = fmt.Sprintf("%s-create-cluster", cluster.Name)
	} else if action == AddNodes {
		name = fmt.Sprintf("%s-add-nodes", cluster.Name)
	}

	podlist := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace("kubekey-system"),
		client.MatchingLabels{"job-name": name},
	}
	for i := 0; i < 10; i++ {
		_ = r.List(context.TODO(), podlist, listOpts...)
		if len(podlist.Items) != 0 {
			if len(podlist.Items[0].ObjectMeta.GetName()) != 0 && len(podlist.Items[0].Status.ContainerStatuses[0].Name) != 0 {
				cluster.Status.JobInfo = kubekeyv1alpha1.JobInfo{
					Namespace: "kubekey-system",
					Name:      name,
					Pods: []kubekeyv1alpha1.PodInfo{{
						Name:       podlist.Items[0].ObjectMeta.GetName(),
						Containers: []kubekeyv1alpha1.ContainerInfo{{Name: podlist.Items[0].Status.ContainerStatuses[0].Name}},
					}},
				}

				if err := r.Status().Update(context.TODO(), cluster); err != nil {
					return err
				}

				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func createClusterConfigMap(r *ClusterReconciler, cluster *kubekeyv1alpha1.Cluster, log logr.Logger) error {
	cmFound := &corev1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: "kubekey-system"}, cmFound); err != nil && kubeErr.IsNotFound(err) {
		// Define a new configmap
		cmCluster := r.configMapForCluster(cluster)
		log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
		if err := r.Create(context.TODO(), cmCluster); err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to create new ConfigMap, ConfigMap.Namespace: %s, ConfigMap.Name: %s", cmCluster.Namespace, cmCluster.Name))
		}
	} else if err != nil {
		return errors.Wrap(err, "Failed to get ConfigMap")
	} else {
		if err1 := r.Delete(context.TODO(), cmFound); err1 != nil {
			return err
		}
		cmCluster := r.configMapForCluster(cluster)
		log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
		if err := r.Create(context.TODO(), cmCluster); err != nil {
			log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", cmCluster.Namespace, "ConfigMap.Name", cmCluster.Name)
			return errors.Wrapf(err, fmt.Sprintf("Failed to create new ConfigMap, ConfigMap.Namespace: %s, ConfigMap.Name: %s", cmCluster.Namespace, cmCluster.Name))
		}
	}
	return nil
}
