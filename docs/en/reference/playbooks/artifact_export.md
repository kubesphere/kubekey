# Export Offline Package (artifact_export.yaml)

`artifact_export.yaml` is used to export a complete offline installation package (artifact) for deploying clusters in air-gapped environments.

## Execution Flow

1. **Global Initialization**
   - Execute the `native/root` role on all nodes (with the `package` tag).

2. **Load Default Variables**
   - Load the `defaults` role on all nodes.

3. **Download and Package**
   - Execute on `localhost`:
     - `download` (with the `package` tag): Download binary files, images, and other resources.
     - `download/package` (with the `package` tag): Package downloaded resources into an offline installation package.

## Notes

When executing this playbook, ensure that the required component versions and image lists have been configured so that the packaged content is complete and usable.
