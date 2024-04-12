/*
 Copyright 2021 The KubeSphere Authors.

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

package images

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	manifestregistry "github.com/estesp/manifest-tool/v2/pkg/registry"
	manifesttypes "github.com/estesp/manifest-tool/v2/pkg/types"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	kubekeyv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	coreutil "github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/registry"
)

type PullImage struct {
	common.KubeAction
}

func (p *PullImage) Execute(runtime connector.Runtime) error {
	if !p.KubeConf.Arg.SkipPullImages {
		i := Images{}
		i.Images = []Image{
			GetImage(runtime, p.KubeConf, "etcd"),
			GetImage(runtime, p.KubeConf, "pause"),
			GetImage(runtime, p.KubeConf, "kube-apiserver"),
			GetImage(runtime, p.KubeConf, "kube-controller-manager"),
			GetImage(runtime, p.KubeConf, "kube-scheduler"),
			GetImage(runtime, p.KubeConf, "kube-proxy"),
			GetImage(runtime, p.KubeConf, "coredns"),
			GetImage(runtime, p.KubeConf, "k8s-dns-node-cache"),
			GetImage(runtime, p.KubeConf, "calico-kube-controllers"),
			GetImage(runtime, p.KubeConf, "calico-cni"),
			GetImage(runtime, p.KubeConf, "calico-node"),
			GetImage(runtime, p.KubeConf, "calico-flexvol"),
			GetImage(runtime, p.KubeConf, "cilium"),
			GetImage(runtime, p.KubeConf, "cilium-operator-generic"),
			GetImage(runtime, p.KubeConf, "flannel"),
			GetImage(runtime, p.KubeConf, "flannel-cni-plugin"),
			GetImage(runtime, p.KubeConf, "kubeovn"),
			GetImage(runtime, p.KubeConf, "haproxy"),
			GetImage(runtime, p.KubeConf, "kubevip"),
		}

		if err := i.PullImages(runtime, p.KubeConf); err != nil {
			return err
		}
	}
	return nil
}

// GetImage defines the list of all images and gets image object by name.
func GetImage(runtime connector.ModuleRuntime, kubeConf *common.KubeConf, name string) Image {
	var image Image
	pauseTag, corednsTag := "3.2", "1.6.9"

	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).LessThan(versionutil.MustParseSemantic("v1.21.0")) {
		pauseTag = "3.2"
		corednsTag = "1.6.9"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.21.0")) ||
		(kubeConf.Cluster.Kubernetes.ContainerManager != "" && kubeConf.Cluster.Kubernetes.ContainerManager != "docker") {
		pauseTag = "3.4.1"
		corednsTag = "1.8.0"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.22.0")) {
		pauseTag = "3.5"
		corednsTag = "1.8.0"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.23.0")) {
		pauseTag = "3.6"
		corednsTag = "1.8.6"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.24.0")) {
		pauseTag = "3.7"
		corednsTag = "1.8.6"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.25.0")) {
		pauseTag = "3.8"
		corednsTag = "1.9.3"
	}
	if versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.26.0")) {
		pauseTag = "3.9"
		corednsTag = "1.9.3"
	}

	logger.Log.Debugf("pauseTag: %s, corednsTag: %s", pauseTag, corednsTag)

	ImageList := map[string]Image{
		"pause":                   {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "pause", Tag: pauseTag, Group: kubekeyv1alpha2.K8s, Enable: true},
		"etcd":                    {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "etcd", Tag: kubekeyv1alpha2.DefaultEtcdVersion, Group: kubekeyv1alpha2.Master, Enable: strings.EqualFold(kubeConf.Cluster.Etcd.Type, kubekeyv1alpha2.Kubeadm)},
		"kube-apiserver":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-apiserver", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-controller-manager": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-controller-manager", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-scheduler":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-scheduler", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-proxy":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-proxy", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.K8s, Enable: !kubeConf.Cluster.Kubernetes.DisableKubeProxy},

		// network
		"coredns":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: corednsTag, Group: kubekeyv1alpha2.K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "k8s-dns-node-cache", Tag: "1.22.20", Group: kubekeyv1alpha2.K8s, Enable: kubeConf.Cluster.Kubernetes.EnableNodelocaldns()},
		"calico-kube-controllers": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-typha":            {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "typha", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico") && len(runtime.GetHostsByRole(common.K8s)) > 50},
		"flannel":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "flannel", Repo: "flannel", Tag: kubekeyv1alpha2.DefaultFlannelVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "flannel")},
		"flannel-cni-plugin":      {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "flannel", Repo: "flannel-cni-plugin", Tag: kubekeyv1alpha2.DefaultFlannelCniPluginVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "flannel")},
		"cilium":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "cilium", Tag: kubekeyv1alpha2.DefaultCiliumVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"cilium-operator-generic": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "operator-generic", Tag: kubekeyv1alpha2.DefaultCiliumVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"hybridnet":               {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "hybridnetdev", Repo: "hybridnet", Tag: kubekeyv1alpha2.DefaulthybridnetVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "hybridnet")},
		"kubeovn":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "kubeovn", Repo: "kube-ovn", Tag: kubekeyv1alpha2.DefaultKubeovnVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "kubeovn")},
		"multus":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "multus-cni", Tag: kubekeyv1alpha2.DefalutMultusVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.Contains(kubeConf.Cluster.Network.Plugin, "multus")},
		// storage
		"provisioner-localpv": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "provisioner-localpv", Tag: "3.3.0", Group: kubekeyv1alpha2.Worker, Enable: false},
		"linux-utils":         {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "linux-utils", Tag: "3.3.0", Group: kubekeyv1alpha2.Worker, Enable: false},
		// load balancer
		"haproxy": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "library", Repo: "haproxy", Tag: "2.9.6-alpine", Group: kubekeyv1alpha2.Worker, Enable: kubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		"kubevip": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "plndr", Repo: "kube-vip", Tag: "v0.7.2", Group: kubekeyv1alpha2.Master, Enable: kubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
		// kata-deploy
		"kata-deploy": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kata-deploy", Tag: "stable", Group: kubekeyv1alpha2.Worker, Enable: kubeConf.Cluster.Kubernetes.EnableKataDeploy()},
		// node-feature-discovery
		"node-feature-discovery": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "node-feature-discovery", Tag: "v0.10.0", Group: kubekeyv1alpha2.K8s, Enable: kubeConf.Cluster.Kubernetes.EnableNodeFeatureDiscovery()},
	}

	image = ImageList[name]
	if kubeConf.Cluster.Registry.NamespaceOverride != "" {
		image.NamespaceOverride = kubeConf.Cluster.Registry.NamespaceOverride
	}
	return image
}

type SaveImages struct {
	common.ArtifactAction
}

func (s *SaveImages) Execute(runtime connector.Runtime) error {
	auths := registry.DockerRegistryAuthEntries(s.Manifest.Spec.ManifestRegistry.Auths)

	dirName := filepath.Join(runtime.GetWorkDir(), common.Artifact, "images")
	if err := coreutil.Mkdir(dirName); err != nil {
		return errors.Wrapf(errors.WithStack(err), "mkdir %s failed", dirName)
	}
	for _, image := range s.Manifest.Spec.Images {
		if err := validateImageName(image); err != nil {
			return err
		}
		imageFullName := strings.Split(image, "/")
		repo := imageFullName[0]
		auth := new(registry.DockerRegistryEntry)
		if v, ok := auths[repo]; ok {
			auth = v
		}

		srcName := fmt.Sprintf("docker://%s", image)
		for _, platform := range s.Manifest.Spec.Arches {
			arch, variant := ParseArchVariant(platform)
			// placeholder
			if variant != "" {
				variant = "-" + variant
			}
			// Ex:
			// oci:./kubekey/artifact/images:kubesphere:kube-apiserver:v1.21.5-amd64
			// oci:./kubekey/artifact/images:kubesphere:kube-apiserver:v1.21.5-arm-v7
			destName := fmt.Sprintf("oci:%s:%s:%s-%s%s", dirName, imageFullName[1], suffixImageName(imageFullName[2:]), arch, variant)
			logger.Log.Infof("Source: %s", srcName)
			logger.Log.Infof("Destination: %s", destName)

			o := &CopyImageOptions{
				srcImage: &srcImageOptions{
					imageName: srcName,
					dockerImage: dockerImageOptions{
						arch:           arch,
						variant:        variant,
						os:             "linux",
						username:       auth.Username,
						password:       auth.Password,
						SkipTLSVerify:  auth.SkipTLSVerify,
						dockerCertPath: auth.CertsPath,
					},
				},
				destImage: &destImageOptions{
					imageName: destName,
					dockerImage: dockerImageOptions{
						arch:    arch,
						variant: variant,
						os:      "linux",
					},
				},
			}

			// Copy image
			// retry 3 times
			for i := 0; i < 3; i++ {
				if err := o.Copy(); err != nil {
					if i == 2 {
						return errors.Wrapf(err, "copy image %s failed", srcName)
					}
					logger.Log.Warnf("copy image %s failed, retrying", srcName)
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
		}
	}
	return nil
}

type CopyImagesToRegistry struct {
	common.KubeAction
	ImagesPath string
}

func (c *CopyImagesToRegistry) Execute(runtime connector.Runtime) error {
	var imagesPath string
	if c.ImagesPath != "" {
		imagesPath = c.ImagesPath
	} else {
		imagesPath = filepath.Join(runtime.GetWorkDir(), "images")
	}

	indexFile, err := os.ReadFile(filepath.Join(imagesPath, "index.json"))
	if err != nil {
		return errors.Errorf("read index.json failed: %s", err)
	}

	index := NewIndex()
	if err := json.Unmarshal(indexFile, index); err != nil {
		return errors.Wrap(errors.WithStack(err), "unmarshal index.json failed: %s")
	}

	auths := registry.DockerRegistryAuthEntries(c.KubeConf.Cluster.Registry.Auths)

	manifestList := make(map[string][]manifesttypes.ManifestEntry)
	for _, m := range index.Manifests {
		ref := m.Annotations.RefName

		// Ex:
		// calico:cni:v3.20.0-amd64
		nameArr := strings.Split(ref, ":")
		if len(nameArr) != 3 {
			return errors.Errorf("invalid ref name: %s", ref)
		}

		image := Image{
			RepoAddr:          c.KubeConf.Cluster.Registry.PrivateRegistry,
			Namespace:         nameArr[0],
			NamespaceOverride: c.KubeConf.Cluster.Registry.NamespaceOverride,
			Repo:              nameArr[1],
			Tag:               nameArr[2],
		}

		uniqueImage, p := ParseImageWithArchTag(image.ImageName())
		entry := manifesttypes.ManifestEntry{
			Image:    image.ImageName(),
			Platform: p,
		}

		skip := false
		if v, ok := manifestList[uniqueImage]; ok {
			// skip if the image already copied
			for _, old := range v {
				if reflect.DeepEqual(old, entry) {
					skip = true
					break
				}
			}

			if !skip {
				v = append(v, entry)
				manifestList[uniqueImage] = v
			}
		} else {
			entryArr := make([]manifesttypes.ManifestEntry, 0)
			manifestList[uniqueImage] = append(entryArr, entry)
		}

		auth := new(registry.DockerRegistryEntry)
		if config, ok := auths[c.KubeConf.Cluster.Registry.GetHost()]; ok {
			auth = config
		}

		srcName := fmt.Sprintf("oci:%s:%s", imagesPath, ref)
		destName := fmt.Sprintf("docker://%s", image.ImageName())
		logger.Log.Infof("Source: %s", srcName)
		logger.Log.Infof("Destination: %s", destName)

		o := &CopyImageOptions{
			srcImage: &srcImageOptions{
				imageName: srcName,
				dockerImage: dockerImageOptions{
					arch:    p.Architecture,
					variant: p.Variant,
					os:      "linux",
				},
			},
			destImage: &destImageOptions{
				imageName: destName,
				dockerImage: dockerImageOptions{
					arch:           p.Architecture,
					variant:        p.Variant,
					os:             "linux",
					username:       auth.Username,
					password:       auth.Password,
					SkipTLSVerify:  auth.SkipTLSVerify,
					dockerCertPath: auth.CertsPath,
				},
			},
		}

		retry, maxRetry := 0, 5
		for ; retry < maxRetry; retry++ {
			if err := o.Copy(); err == nil {
				break
			} else {
				fmt.Println(errors.WithStack(err))
			}
		}
		if retry >= maxRetry {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("copy image %s to %s failed, retry %d", srcName, destName, maxRetry))
		}
	}

	c.ModuleCache.Set("manifestList", manifestList)

	return nil
}

type PushManifest struct {
	common.KubeAction
}

func (p *PushManifest) Execute(_ connector.Runtime) error {
	// make a multi-arch image
	// push a manifest list to the private registry.

	v, ok := p.ModuleCache.Get("manifestList")
	if !ok {
		return errors.New("get manifest list failed by module cache")
	}
	list := v.(map[string][]manifesttypes.ManifestEntry)

	auths := registry.DockerRegistryAuthEntries(p.KubeConf.Cluster.Registry.Auths)
	auth := new(registry.DockerRegistryEntry)
	if _, ok := auths[p.KubeConf.Cluster.Registry.PrivateRegistry]; ok {
		auth = auths[p.KubeConf.Cluster.Registry.PrivateRegistry]
	}

	for imageName, platforms := range list {
		manifestSpec := NewManifestSpec(imageName, platforms)
		logger.Log.Debug(manifestSpec)

		logger.Log.Infof("Push multi-arch manifest list: %s", imageName)
		// todo: the function can't support specify a certs dir
		digest, length, err := manifestregistry.PushManifestList(auth.Username, auth.Password, manifestSpec,
			true, true, auth.PlainHTTP, "")
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("push image %s multi-arch manifest failed", imageName))
		}
		logger.Log.Infof("Digest: %s Length: %d", digest, length)
	}

	return nil
}
