# Initialize Registry (init_registry.yaml)

`init_registry.yaml` is used to initialize and deploy a private image registry (such as Harbor) on specified nodes, including pre-checks, resource download, certificate generation, and registry installation.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes.

2. **Load Default Variables and Precheck**
   - Load the `defaults` role on all nodes.
   - Execute `precheck/image-registry` to check whether image registry nodes meet the installation conditions.

3. **Certificate and Resource Preparation**
   - Execute on `localhost`:
     - `certs/init`: Generate certificates.
     - `download`: Download image registry-related software packages and images.

4. **Install Image Registry**
   - Execute in sequence for nodes in the `image_registry` group:
     - `native/init`: Initialize the system environment.
     - `native/dns`: Configure local DNS resolution.
     - `image-registry`: Install and configure the private image registry.

## Notes

- Currently supported image registry types include Harbor and docker-registry, determined by `image_registry.type`.
- If using Harbor, please ensure that `docker_version` and `dockercompose_version` are correctly configured.
