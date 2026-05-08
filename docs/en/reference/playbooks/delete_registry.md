# Delete Registry (delete_registry.yaml)

`delete_registry.yaml` is used to uninstall a deployed private image registry (such as Harbor or docker-registry) and clean up local DNS resolution configuration.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.
   - Load the `defaults` role on all nodes.

2. **Uninstall Image Registry**
   - Execute `uninstall/image-registry` for nodes in the `image_registry` group.

3. **Clean Local DNS Configuration**
   - Clean up locally-written DNS (hosts) marker segments on `image_registry` nodes.
   - Triggered only when `delete.dns` is `true`.

## Notes

- Before deleting the image registry, please ensure that there are no images in the registry that the cluster depends on, or that images have been migrated.
- This operation only uninstalls the image registry service; it does not delete the container runtime or Kubernetes components in the cluster.
