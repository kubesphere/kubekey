# Sync Offline Images (artifact_images.yaml)

`artifact_images.yaml` is used to pull container images required by the cluster and push them to a private image registry. It is commonly used for image preparation in offline environments.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.

2. **Load Default Variables**
   - Load the `defaults` role on `localhost`.

3. **Image Download and Push**
   - Execute in sequence on `localhost`:
     - `download`: Download required resources.
     - `image-registry/pull` (with the `pull` tag): Pull container images locally.
     - `image-registry/push` (with the `push` tag): Push pulled images to the configured private image registry.

## Notes

- This playbook is typically executed on `localhost`.
- The push target registry is determined by the `image_registry` configuration.
