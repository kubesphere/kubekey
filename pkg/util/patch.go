package util

import (
	"context"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// PatchHelper is a utility struct that provides methods to help with patching Kubernetes objects.
// It contains a client for interacting with the Kubernetes API server,
// and a map of helpers for different GroupVersionKinds to facilitate patch operations.
type PatchHelper struct {
	client  ctrlclient.Client
	helpers map[schema.GroupVersionKind]*patch.Helper
}

// NewPatchHelper creates a new PatchHelper instance for the given runtime client.
// It accepts a variable number of ctrlclient.Object and creates a patch.Helper for each object.
// The function returns a pointer to a PatchHelper and an error if any occurs during the creation of patch helpers.
//
// Parameters:
// - client: The ctrlclient.Client to be used for creating patch helpers.
// - obj: A variadic parameter of ctrlclient.Object for which patch helpers will be created.
//
// Returns:
// - *PatchHelper: A pointer to the created PatchHelper instance.
// - error: An error if any occurs during the creation of patch helpers.
func NewPatchHelper(client ctrlclient.Client, obj ...ctrlclient.Object) (*PatchHelper, error) {
	helpers := make(map[schema.GroupVersionKind]*patch.Helper)
	for _, o := range obj {
		if o == nil || o.GetName() == "" {
			continue
		}
		helper, err := patch.NewHelper(o, client)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create patch helper for object %q", ctrlclient.ObjectKeyFromObject(o))
		}
		gvk, err := apiutil.GVKForObject(o, client.Scheme())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get gvk for object %q", ctrlclient.ObjectKeyFromObject(o))
		}
		helpers[gvk] = helper
	}

	return &PatchHelper{
		client:  client,
		helpers: helpers,
	}, nil
}

// Patch applies the patch operation to the provided objects.
// It iterates over the given objects, determines their GroupVersionKind (GVK),
// and applies the corresponding patch helper for each object.
//
// Parameters:
//
//	ctx - The context for the patch operation.
//	obj - A variadic list of ctrlclient.Object to be patched.
//
// Returns:
//
//	An error if any of the patch operations fail, otherwise nil.
func (p *PatchHelper) Patch(ctx context.Context, obj ...ctrlclient.Object) error {
	for _, o := range obj {
		gvk, err := apiutil.GVKForObject(o, p.client.Scheme())
		if err != nil {
			return errors.Wrapf(err, "failed to get gvk for object %q", ctrlclient.ObjectKeyFromObject(o))
		}
		if p.helpers[gvk] == nil {
			// object is created, should not patch.
			return nil
		}
		if err := p.helpers[gvk].Patch(ctx, o); err != nil {
			return errors.Wrapf(err, "failed to patch object %q", ctrlclient.ObjectKeyFromObject(o))
		}
	}

	return nil
}
