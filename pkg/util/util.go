package util

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// GetOwnerFromObject retrieves the owner object of a given resource by inspecting the OwnerReferences field.
// It searches for an owner that matches the specified type and group.
// If the owner is found, it populates the 'owner' parameter with the retrieved object.
//
// Parameters:
// - ctx: Context for the API request.
// - scheme: The runtime scheme used to determine the GroupVersionKind (GVK) of the owner object.
// - client: The Kubernetes client to fetch resources.
// - obj: The resource object whose owner is being queried.
// - owner: An empty object of the desired owner type that will be populated upon success.
//
// Returns:
// - error: An error if the owner cannot be found or fetched.
func GetOwnerFromObject(ctx context.Context, scheme *runtime.Scheme, client ctrlclient.Client, obj ctrlclient.Object, owner ctrlclient.Object) error {
	gvk, err := apiutil.GVKForObject(owner, scheme)
	if err != nil {
		return err
	}

	for _, ref := range obj.GetOwnerReferences() {
		// Check if the owner reference matches the expected Kind and Group
		if ref.Kind != reflect.TypeOf(owner).Name() {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return err
		}
		if gv.Group == gvk.Group {
			// Retrieve the owner object using the Kubernetes client
			return client.Get(ctx, ctrlclient.ObjectKey{Name: ref.Name, Namespace: obj.GetNamespace()}, owner)
		}
	}

	return fmt.Errorf("owner %s not found", gvk.String())
}

// GetObjectListFromOwner filters a list of Kubernetes objects to include only those owned by a specified resource.
// It inspects each object's OwnerReferences to find a match based on the specified owner's Group and Kind.
//
// Parameters:
// - ctx: Context for the API request.
// - scheme: The runtime scheme used to determine the GroupVersionKind (GVK) of the owner object.
// - client: The Kubernetes client to list and fetch resources.
// - objList: The empty list of objects to be filtered (must implement ctrlclient.ObjectList).
// - owner: The owner object to match against.
//
// Returns:
// - error: An error if the list cannot be retrieved, parsed, or filtered.
//
// Behavior:
// - Filters objList.Items to retain only objects whose OwnerReferences match the provided owner's Group and Kind.
// - Updates objList.Items in place to include only matching items.
func GetObjectListFromOwner(ctx context.Context, scheme *runtime.Scheme, client ctrlclient.Client, owner ctrlclient.Object, objList ctrlclient.ObjectList, opts ...ctrlclient.ListOption) error {
	// Retrieve the GroupVersionKind (GVK) of the owner object
	gvk, err := apiutil.GVKForObject(owner, scheme)
	if err != nil {
		return err
	}
	opts = append(opts, ctrlclient.InNamespace(owner.GetNamespace()))
	// List all objects in the namespace of the owner
	if err := client.List(ctx, objList, opts...); err != nil {
		return err
	}

	// Access the Items field of objList using reflection
	objListValue := reflect.ValueOf(objList).Elem()
	itemsField := objListValue.FieldByName("Items")
	if !itemsField.IsValid() || itemsField.Kind() != reflect.Slice {
		return errors.New("objList does not have a valid Items field")
	}

	// Create a slice to hold objects matching the owner's reference
	filteredItems := reflect.MakeSlice(itemsField.Type(), 0, itemsField.Len())

	// Iterate through the list and filter objects by OwnerReferences
	for i := range itemsField.Len() {
		item, ok := itemsField.Index(i).Addr().Interface().(ctrlclient.Object)
		if !ok {
			return errors.New("item does not implement ctrlclient.Object")
		}
		matched := false

		// Check OwnerReferences for a match
		for _, ref := range item.GetOwnerReferences() {
			if ref.Kind != reflect.TypeOf(owner).Name() {
				continue
			}
			gv, err := schema.ParseGroupVersion(ref.APIVersion)
			if err != nil {
				return err
			}
			if gv.Group == gvk.Group {
				matched = true

				break
			}
		}

		// Append the item to the filtered list if it matches
		if matched {
			filteredItems = reflect.Append(filteredItems, itemsField.Index(i))
		}
	}

	// Update objList.Items with the filtered objects
	itemsField.Set(filteredItems)

	return nil
}

// ObjectRef creates an ObjectReference for the given Kubernetes object.
// It takes a runtime.Scheme and a ctrlclient.Object as parameters and returns
// a pointer to a corev1.ObjectReference. If it fails to get the GroupVersionKind (GVK)
// of the object, it logs an error and returns nil.
//
// Parameters:
//   - scheme: A pointer to a runtime.Scheme used to get the GVK of the object.
//   - obj: A ctrlclient.Object for which the ObjectReference is created.
//
// Returns:
//   - A pointer to a corev1.ObjectReference containing the Kind, APIVersion, Name,
//     and Namespace of the given object, or nil if an error occurs.
func ObjectRef(scheme *runtime.Scheme, obj ctrlclient.Object) *corev1.ObjectReference {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		klog.ErrorS(err, "failed to get GVK", "object", ctrlclient.ObjectKeyFromObject(obj))

		return nil
	}

	return &corev1.ObjectReference{
		Kind:       gvk.Kind,
		APIVersion: gvk.GroupVersion().String(),
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}
}
