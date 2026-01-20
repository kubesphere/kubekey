package v1beta1

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/emicklei/go-restful/v3"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/web/api"
	"github.com/kubesphere/kubekey/v4/pkg/web/query"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler handle cluster-api resources
type Handler struct {
	workDir string
	client  ctrlclient.Client
	config  *rest.Config
}

// NewHandler create a new Handler
func NewHandler(client ctrlclient.Client, config *rest.Config, workDir string) *Handler {
	return &Handler{
		workDir: workDir,
		client:  client,
		config:  config,
	}
}

// CreateKKCluster create a kk cluster resource
// one namespace has only one cluster, so namespace should equal with kk-cluster-name
// may i delete namespace ,set KKCluster resource scope=Cluster?
// also should create Inventory and KKMachineTemplate and KKMachineList
// create Inventory as a middle status data ,record all node hosts
// create KKMachineTemplate for cluster configs and all nodes
// create KKMachineList for all nodes
func (h *Handler) CreateKKCluster(req *restful.Request, resp *restful.Response) {
	var kkcluster v1beta1.KKCluster
	err := req.ReadEntity(&kkcluster)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}
	err = h.client.Create(req.Request.Context(), &kkcluster)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	hosts, err := converter.ConvertKKClusterToInventoryHost(&kkcluster)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	err = h.createInventoryWithKKCluster(req.Request.Context(), kkcluster, hosts)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	err = h.createKKMachineListWithKKCluster(req.Request.Context(), kkcluster, hosts)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	resp.WriteEntity(kkcluster)
}

// DeleteKKCluster delete kk cluster
func (h *Handler) DeleteKKCluster(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	kkClusterName := req.PathParameter("kk-cluster-name")

	err := h.client.Delete(req.Request.Context(), &v1beta1.KKCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kkClusterName,
			Namespace: namespace,
		},
	})

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	resp.WriteEntity(v1beta1.KKCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kkClusterName,
			Namespace: namespace,
		},
	})
}

// CreateKKMachine create a kk machine , only used when kk cluster created
// one kk machine equals one node
// after create kk machine , also need update kk cluster inventory
func (h *Handler) CreateKKMachine(req *restful.Request, resp *restful.Response) {
	var kkmachine v1beta1.KKMachine
	err := req.ReadEntity(&kkmachine)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	if kkmachine.Labels == nil {
		kkmachine.Labels = make(map[string]string)
	}

	if _, ok := kkmachine.Labels[_const.ClusterApiClusterNameLabelKey]; !ok {
		kkmachine.Labels[_const.ClusterApiClusterNameLabelKey] = kkmachine.GetNamespace()
	}

	clusterName := kkmachine.Labels[_const.ClusterApiClusterNameLabelKey]

	var inventory kkcorev1.Inventory
	err = h.client.Get(req.Request.Context(), ctrlclient.ObjectKey{
		Name:      clusterName,
		Namespace: kkmachine.GetNamespace(),
	}, &inventory)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	err = h.client.Create(req.Request.Context(), &kkmachine)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	err = h.updateInventoryHostGroupWithNewMachine(req.Request.Context(), inventory, kkmachine)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	resp.WriteEntity(kkmachine)
}

// DeleteKKMachine delete kk machine
func (h *Handler) DeleteKKMachine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	kkMachineName := req.PathParameter("kk-machine-name")
	kkClusterName := req.QueryParameter("kk-cluster-name")

	err := h.client.Delete(req.Request.Context(), &v1beta1.KKMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kkMachineName,
			Namespace: namespace,
		},
	})
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}
	err = h.clearInventoryAfterDeleteKKMachine(req.Request.Context(), namespace, kkClusterName, kkMachineName)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	resp.WriteEntity(v1beta1.KKMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kkMachineName,
			Namespace: namespace,
			Labels: map[string]string{
				_const.ClusterApiClusterNameLabelKey: kkClusterName,
			},
		},
	})
}

func (h *Handler) UpdateKKMachine(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	kkMachineName := req.PathParameter("kk-machine-name")
	kkClusterName := req.QueryParameter("kk-cluster-name")

	var tmpl = v1beta1.KKMachine{}
	err := h.client.Get(req.Request.Context(), ctrlclient.ObjectKey{
		Name:      kkMachineName,
		Namespace: namespace,
	}, &tmpl)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	if machineClusterName, ok := tmpl.GetLabels()[_const.ClusterApiClusterNameLabelKey]; ok && machineClusterName != kkClusterName {
		err = errors.Newf("kk machine not belone to input cluster")
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	contentType := req.HeaderParameter("Content-Type")
	patchType := types.PatchType(contentType)

	patchBody, err := io.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleError(resp, req, errors.Wrap(err, "failed to read patch body from request"))
		return
	}
	codec := _const.CodecFactory.LegacyCodec(v1beta1.SchemeGroupVersion)

	oldItemJson, err := runtime.Encode(codec, &tmpl)
	if err != nil {
		api.HandleError(resp, req, errors.Wrap(err, "failed to encode old inventory object to JSON"))
		return
	}

	applyPatchAndDecode := func(objectJSON []byte) (*v1beta1.KKMachine, error) {
		var patchedJSON []byte
		switch patchType {
		case types.JSONPatchType:
			patchObj, err := jsonpatch.DecodePatch(patchBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decode JSON patch")
			}
			patchedJSON, err = patchObj.Apply(objectJSON)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply JSON patch to old inventory JSON")
			}
		case types.MergePatchType:
			var err error
			patchedJSON, err = jsonpatch.MergePatch(objectJSON, patchBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply merge patch to old inventory JSON")
			}
		default:
			return nil, errors.Errorf("unknown Content-Type header for patch: %v", patchType)
		}
		newItem := &v1beta1.KKMachine{}
		err := runtime.DecodeInto(codec, patchedJSON, newItem)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode patched inventory JSON")
		}
		return newItem, nil
	}

	updatedItem, err := applyPatchAndDecode(oldItemJson)
	if err != nil {
		api.HandleError(resp, req, errors.Wrap(err, "failed to apply patch and decode inventory"))
		return
	}

	if err = h.client.Patch(req.Request.Context(), updatedItem, ctrlclient.MergeFrom(&tmpl)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(updatedItem)

}

// CreateConfig update KKMachineTemplate
// after create KKCluster, KKMachineTemplate will be created with out config
// so we set config by this function, patch KKMachineTemplate and set config
func (h *Handler) CreateConfig(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	kkClusterName := req.PathParameter("kk-cluster-name")

	var tmpl = v1beta1.KKCluster{}
	err := h.client.Get(req.Request.Context(), ctrlclient.ObjectKey{
		Name:      kkClusterName,
		Namespace: namespace,
	}, &tmpl)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	contentType := req.HeaderParameter("Content-Type")
	patchType := types.PatchType(contentType)

	patchBody, err := io.ReadAll(req.Request.Body)
	if err != nil {
		api.HandleError(resp, req, errors.Wrap(err, "failed to read patch body from request"))
		return
	}
	codec := _const.CodecFactory.LegacyCodec(v1beta1.SchemeGroupVersion)

	oldItemJson, err := runtime.Encode(codec, &tmpl)
	if err != nil {
		api.HandleError(resp, req, errors.Wrap(err, "failed to encode old inventory object to JSON"))
		return
	}

	applyPatchAndDecode := func(objectJSON []byte) (*v1beta1.KKCluster, error) {
		var patchedJSON []byte
		switch patchType {
		case types.JSONPatchType:
			patchObj, err := jsonpatch.DecodePatch(patchBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to decode JSON patch")
			}
			patchedJSON, err = patchObj.Apply(objectJSON)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply JSON patch to old inventory JSON")
			}
		case types.MergePatchType:
			var err error
			patchedJSON, err = jsonpatch.MergePatch(objectJSON, patchBody)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply merge patch to old inventory JSON")
			}
		default:
			return nil, errors.Errorf("unknown Content-Type header for patch: %v", patchType)
		}
		newItem := &v1beta1.KKCluster{}
		err := runtime.DecodeInto(codec, patchedJSON, newItem)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode patched inventory JSON")
		}
		return newItem, nil
	}

	updatedItem, err := applyPatchAndDecode(oldItemJson)
	if err != nil {
		api.HandleError(resp, req, errors.Wrap(err, "failed to apply patch and decode inventory"))
		return
	}

	if err = h.client.Patch(req.Request.Context(), updatedItem, ctrlclient.MergeFrom(&tmpl)); err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// do not return host info after cluster patch
	updatedItem.Spec.InventoryHosts = nil

	resp.WriteEntity(updatedItem)
}

// GetKKCluster query one kk cluster data with config
func (h *Handler) GetKKCluster(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	kkClusterName := req.PathParameter("kk-cluster-name")

	var kkCluster v1beta1.KKCluster
	err := h.client.Get(req.Request.Context(), ctrlclient.ObjectKey{
		Name:      kkClusterName,
		Namespace: namespace,
	}, &kkCluster)
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}
	// do not return host info by cluster query
	kkCluster.Spec.InventoryHosts = nil

	resp.WriteEntity(kkCluster)
}

// GetKKMachineList get KKMachine list
// query node list
func (h *Handler) GetKKMachineList(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	kkClusterName := req.PathParameter("kk-cluster-name")

	queryParam := query.ParseQueryParameter(req)
	var fieldselector fields.Selector
	// Parse field selector from query parameters if present.
	if v, ok := queryParam.Filters[query.ParameterFieldSelector]; ok {
		fs, err := fields.ParseSelector(v)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		fieldselector = fs
	}

	rqm, _ := labels.NewRequirement(_const.ClusterApiClusterNameLabelKey, selection.Equals, []string{kkClusterName})

	var items v1beta1.KKMachineList

	err := h.client.List(req.Request.Context(), &items, &ctrlclient.ListOptions{
		Namespace:     namespace,
		LabelSelector: queryParam.Selector().Add(*rqm),
		FieldSelector: fieldselector,
	})
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	result := query.DefaultList(items.Items, queryParam, func(left, right v1beta1.KKMachine, sortBy string) bool {
		leftMeta, err := meta.Accessor(left)
		if err != nil {
			return false
		}
		rightMeta, err := meta.Accessor(right)
		if err != nil {
			return false
		}

		return query.DefaultObjectMetaCompare(leftMeta, rightMeta, sortBy)
	}, func(o v1beta1.KKMachine, filter query.Filter) bool {
		// Skip fieldselector filter.
		if filter.Field == query.ParameterFieldSelector {
			return true
		}
		objectMeta, err := meta.Accessor(o)
		if err != nil {
			return false
		}

		return query.DefaultObjectMetaFilter(objectMeta, filter)
	})

	resp.WriteEntity(result)
}

// GetKKClusterList get KKCluster list
// get cluster list
func (h *Handler) GetKKClusterList(req *restful.Request, resp *restful.Response) {
	var items v1beta1.KKClusterList

	queryParam := query.ParseQueryParameter(req)
	var fieldselector fields.Selector
	// Parse field selector from query parameters if present.
	if v, ok := queryParam.Filters[query.ParameterFieldSelector]; ok {
		fs, err := fields.ParseSelector(v)
		if err != nil {
			api.HandleError(resp, req, err)
			return
		}
		fieldselector = fs
	}

	err := h.client.List(req.Request.Context(), &items, &ctrlclient.ListOptions{
		LabelSelector: queryParam.Selector(),
		FieldSelector: fieldselector,
	})
	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(resp, req, err)
		return
	}

	result := query.DefaultList(items.Items, queryParam, func(left, right v1beta1.KKCluster, sortBy string) bool {
		leftMeta, err := meta.Accessor(left)
		if err != nil {
			return false
		}
		rightMeta, err := meta.Accessor(right)
		if err != nil {
			return false
		}
		return query.DefaultObjectMetaCompare(leftMeta, rightMeta, sortBy)
	}, func(o v1beta1.KKCluster, filter query.Filter) bool {
		// Skip fieldselector filter.
		if filter.Field == query.ParameterFieldSelector {
			return true
		}
		objectMeta, err := meta.Accessor(o)
		if err != nil {
			return false
		}
		return query.DefaultObjectMetaFilter(objectMeta, filter)
	}, func(cluster v1beta1.KKCluster) v1beta1.KKCluster {
		cluster.Spec.InventoryCount = len(cluster.Spec.InventoryHosts)
		cluster.Spec.InventoryHosts = nil
		return cluster
	})

	resp.WriteEntity(result)

}

func (h *Handler) createInventoryWithKKCluster(ctx context.Context, kkCluster v1beta1.KKCluster, hosts kkcorev1.InventoryHost) error {

	inventory := kkcorev1.Inventory{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kkCluster.Name,
			Namespace: kkCluster.Namespace,
			Labels: map[string]string{
				_const.ClusterApiClusterNameLabelKey: kkCluster.Name,
			},
		},
		Spec: kkcorev1.InventorySpec{
			Hosts:  hosts,
			Groups: convertInventoryHostToGroup(hosts),
		},
	}

	return h.client.Create(ctx, &inventory)
}

func (h *Handler) createKKMachineListWithKKCluster(ctx context.Context, kkCluster v1beta1.KKCluster, hosts kkcorev1.InventoryHost) error {

	for _, host := range kkCluster.Spec.InventoryHosts {
		km := v1beta1.KKMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      host.Name,
				Namespace: kkCluster.Namespace,
				Labels: map[string]string{
					_const.ClusterApiClusterNameLabelKey: kkCluster.Name,
				},
			},
			Spec: v1beta1.KKMachineSpec{
				ProviderID: ptr.To(fmt.Sprintf("kk://%s/%s", kkCluster.Name, host.Name)),
				Roles: []string{
					"control-plane",
					"master",
					"worker",
				},
				Config: hosts[host.Name],
				Taints: host.Taints,
			},
		}
		err := h.client.Create(ctx, &km)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) updateInventoryHostGroupWithNewMachine(ctx context.Context,
	inventory kkcorev1.Inventory, kkMachine v1beta1.KKMachine) error {

	for _, role := range kkMachine.Spec.Roles {
		ig, ok := inventory.Spec.Groups[role]
		if !ok {
			ig = kkcorev1.InventoryGroup{
				Hosts: make([]string, 0),
			}
		}
		groupExists := false
		for _, gh := range ig.Hosts {
			if gh == kkMachine.GetName() {
				groupExists = true
			}
		}
		if !groupExists {
			ig.Hosts = append(ig.Hosts, kkMachine.GetName())
			inventory.Spec.Groups[role] = ig
		}
	}
	inventory.Spec.Hosts[kkMachine.GetName()] = kkMachine.Spec.Config

	return h.client.Update(ctx, &inventory)
}

func (h *Handler) clearInventoryAfterDeleteKKMachine(ctx context.Context, namespace, kkclustername, kkmachinename string) error {
	var inventory kkcorev1.Inventory

	err := h.client.Get(ctx, ctrlclient.ObjectKey{
		Name:      kkclustername,
		Namespace: namespace,
	}, &inventory)

	if err != nil {
		return err
	}

	if _, ok := inventory.Spec.Hosts[kkmachinename]; ok {
		delete(inventory.Spec.Hosts, kkmachinename)
		for groupName, groupData := range inventory.Spec.Groups {
			filteredHost := make([]string, 0)
			for _, host := range groupData.Hosts {
				if host != kkmachinename {
					filteredHost = append(filteredHost, host)
				}
			}
			groupData.Hosts = filteredHost
			inventory.Spec.Groups[groupName] = groupData
		}
		// only need update inventory when machine in inventory
		err = h.client.Update(ctx, &inventory)
	}

	return err
}

func convertInventoryHostToGroup(hosts kkcorev1.InventoryHost) map[string]kkcorev1.InventoryGroup {
	existGroups := make(map[string]kkcorev1.InventoryGroup)
	for hostName, hostVars := range hosts {
		hostData := variable.Extension2Variables(hostVars)
		roles, err := variable.StringSliceVar(map[string]any{}, hostData, "roles")
		if err != nil {
			continue
		}
		for _, role := range roles {
			hs, ok := existGroups[role]
			if !ok {
				hs = kkcorev1.InventoryGroup{
					Hosts: make([]string, 0),
				}
			}
			hs.Hosts = append(hs.Hosts, hostName)
			existGroups[role] = hs
		}
	}
	return existGroups
}
