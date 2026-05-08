# Dependency Packages

KubeKey provides pre-compiled system dependency packages for the following distributions, available at [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest):

- AlmaLinux 9.0
- CentOS 8
- Debian 10 / 11
- Kylin V10SP1 / V10SP2 / V10SP3 / V10SP3-2403
- Ubuntu 18.04 / 20.04 / 22.04 / 24.04

> Note: Some distributions (e.g. CentOS 7, Ubuntu 16.04) have Dockerfiles in the repository, but are not included in the [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest) automated build due to being EOL or no longer officially maintained. Please refer to `.github/workflows/gen-repository-iso.yaml` for the exact build matrix.

## Dependency Package Build Process

The build scripts for each distribution are located in the [`hack/gen-repository-iso`](../../hack/gen-repository-iso) directory. Build artifacts are automatically published to [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest) by `.github/workflows/gen-repository-iso.yaml`.

### File Descriptions

- `dockerfile.<os>`: Dockerfiles for each distribution, used to package dependencies inside a container.
- `download-pkgs.sh`: Generic download script.
- `packages.yaml`: Declares all dependency packages that need to be pre-compiled.

### Manual Build

If you need to customize dependency packages, enter the `hack/gen-repository-iso` directory and execute:

```shell
docker build -f dockerfile.<os> -t kk-iso .
```

Or use Docker Buildx to build multi-arch images and export artifacts:

```shell
docker buildx build --platform linux/amd64,linux/arm64 -f dockerfile.<os> -o ./output .
```

After the build completes, the corresponding `.iso` file can be found in the output directory.
