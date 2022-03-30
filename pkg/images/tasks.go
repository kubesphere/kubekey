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
	manifesttypes "github.com/estesp/manifest-tool/v2/pkg/types"
	"github.com/kubesphere/kubekey/pkg/container/templates"
	coreutil "github.com/kubesphere/kubekey/pkg/core/util"
	"io/ioutil"
	"path/filepath"
	"strings"

	manifestregistry "github.com/estesp/manifest-tool/v2/pkg/registry"
	kubekeyv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type PullImage struct {
	common.KubeAction
}

func (p *PullImage) Execute(runtime connector.Runtime) error {
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
		GetImage(runtime, p.KubeConf, "operator-generic"),
		GetImage(runtime, p.KubeConf, "flannel"),
		GetImage(runtime, p.KubeConf, "kubeovn"),
		GetImage(runtime, p.KubeConf, "haproxy"),
	}
	if err := i.PullImages(runtime, p.KubeConf); err != nil {
		return err
	}
	return nil
}

// GetImage defines the list of all images and gets image object by name.
func GetImage(runtime connector.ModuleRuntime, kubeConf *common.KubeConf, name string) Image {
	var image Image
	var pauseTag, corednsTag string

	cmp, err := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.21.0")
	if err != nil {
		logger.Log.Fatal("Failed to compare version: %v", err)
	}
	if (cmp == 0 || cmp == 1) || (kubeConf.Cluster.Kubernetes.ContainerManager != "" && kubeConf.Cluster.Kubernetes.ContainerManager != "docker") {
		cmp, err := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.22.0")
		if err != nil {
			logger.Log.Fatal("Failed to compare version: %v", err)
		}
		if cmp == 0 || cmp == 1 {
			pauseTag = "3.5"
		} else {
			pauseTag = "3.4.1"
		}
	} else {
		pauseTag = "3.2"
	}
	cmp2, err2 := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.23.0")
	if err2 != nil {
		logger.Log.Fatal("Failed to compare version: %v", err)
	}
	if cmp2 == 0 || cmp2 == 1 {
		pauseTag = "3.6"
	}
	// get coredns image tag
	if cmp == -1 {
		corednsTag = "1.6.9"
	} else {
		corednsTag = "1.8.0"
	}

	ImageList := map[string]Image{
		"pause":                   {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "pause", Tag: pauseTag, Group: kubekeyv1alpha2.K8s, Enable: true},
		"etcd":                    {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "etcd", Tag: kubekeyv1alpha2.DefaultEtcdVersion, Group: kubekeyv1alpha2.Master, Enable: strings.EqualFold(kubeConf.Cluster.Etcd.Type, kubekeyv1alpha2.Kubeadm)},
		"kube-apiserver":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-apiserver", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-controller-manager": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-controller-manager", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-scheduler":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-scheduler", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.Master, Enable: true},
		"kube-proxy":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "kube-proxy", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyv1alpha2.K8s, Enable: true},

		// network
		"coredns":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: corednsTag, Group: kubekeyv1alpha2.K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: kubekeyv1alpha2.K8s, Enable: kubeConf.Cluster.Kubernetes.EnableNodelocaldns()},
		"calico-kube-controllers": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-typha":            {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "typha", Tag: kubekeyv1alpha2.DefaultCalicoVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico") && len(runtime.GetHostsByRole(common.K8s)) > 50},
		"flannel":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "flannel", Tag: kubekeyv1alpha2.DefaultFlannelVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "flannel")},
		"cilium":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "cilium", Tag: kubekeyv1alpha2.DefaultCiliumVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"operator-generic":        {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "operator-generic", Tag: kubekeyv1alpha2.DefaultCiliumVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"kubeovn":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "kubeovn", Repo: "kube-ovn", Tag: kubekeyv1alpha2.DefaultKubeovnVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "kubeovn")},
		"multus":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyv1alpha2.DefaultKubeImageNamespace, Repo: "multus-cni", Tag: kubekeyv1alpha2.DefalutMultusVersion, Group: kubekeyv1alpha2.K8s, Enable: strings.Contains(kubeConf.Cluster.Network.Plugin, "multus")},
		// storage
		"provisioner-localpv": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "provisioner-localpv", Tag: "2.10.1", Group: kubekeyv1alpha2.Worker, Enable: false},
		"linux-utils":         {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "linux-utils", Tag: "2.10.0", Group: kubekeyv1alpha2.Worker, Enable: false},
		// load balancer
		"haproxy": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "library", Repo: "haproxy", Tag: "2.3", Group: kubekeyv1alpha2.Worker, Enable: kubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
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
	auths := Auths(s.Manifest)

	dirName := filepath.Join(runtime.GetWorkDir(), common.Artifact, "images")
	if err := coreutil.Mkdir(dirName); err != nil {
		return errors.Wrapf(errors.WithStack(err), "mkdir %s failed", dirName)
	}
	for _, image := range s.Manifest.Spec.Images {
		imageFullName := strings.Split(image, "/")
		repo := imageFullName[0]
		var registryAuth registryAuth
		if v, ok := auths[repo]; ok {
			registryAuth = v
		}

		srcName := fmt.Sprintf("docker://%s", image)
		for _, platform := range s.Manifest.Spec.Arches {
			arch, variant := ParseArchVariant(platform)
			// placeholder
			if variant != "" {
				variant = "-" + variant
			}
			// Ex:
			// oci:./kubekey/artifact/images:kubesphere:kube-apiserver-amd64:v1.21.5
			// oci:./kubekey/artifact/images:kubesphere:kube-apiserver-arm-v7:v1.21.5
			destName := fmt.Sprintf("oci:%s:%s:%s-%s%s", dirName, imageFullName[1], imageFullName[2], arch, variant)
			logger.Log.Infof("Source: %s", srcName)
			logger.Log.Infof("Destination: %s", destName)

			o := &CopyImageOptions{
				srcImage: &srcImageOptions{
					imageName: srcName,
					dockerImage: dockerImageOptions{
						arch:      arch,
						variant:   variant,
						os:        "linux",
						username:  registryAuth.Username,
						password:  registryAuth.Password,
						tlsVerify: registryAuth.PlainHTTP,
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

			if err := o.Copy(); err != nil {
				return err
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

	indexFile, err := ioutil.ReadFile(filepath.Join(imagesPath, "index.json"))
	if err != nil {
		return errors.Errorf("read index.json failed: %s", err)
	}

	index := NewIndex()
	if err := json.Unmarshal(indexFile, index); err != nil {
		return errors.Wrap(errors.WithStack(err), "unmarshal index.json failed: %s")
	}

	auths := templates.Auths(c.KubeConf)

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
		if v, ok := manifestList[uniqueImage]; ok {
			v = append(v, entry)
			manifestList[uniqueImage] = v
		} else {
			entryArr := make([]manifesttypes.ManifestEntry, 0)
			manifestList[uniqueImage] = append(entryArr, entry)
		}

		var auth templates.DockerConfigEntry
		if config, ok := auths[c.KubeConf.Cluster.Registry.PrivateRegistry]; ok {
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
					arch:      p.Architecture,
					variant:   p.Variant,
					os:        "linux",
					username:  auth.Username,
					password:  auth.Password,
					tlsVerify: c.KubeConf.Cluster.Registry.PlainHTTP,
				},
			},
		}

		if err := o.Copy(); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("copy image %s to %s failed", srcName, destName))
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

	auths := templates.Auths(p.KubeConf)
	var auth templates.DockerConfigEntry
	if _, ok := auths[p.KubeConf.Cluster.Registry.PrivateRegistry]; ok {
		auth = auths[p.KubeConf.Cluster.Registry.PrivateRegistry]
	}

	for imageName, platforms := range list {
		manifestSpec := NewManifestSpec(imageName, platforms)
		logger.Log.Debug(manifestSpec)

		logger.Log.Infof("Push multi-arch manifest list: %s", imageName)
		digest, length, err := manifestregistry.PushManifestList(auth.Username, auth.Password, manifestSpec, false, true,
			p.KubeConf.Cluster.Registry.PlainHTTP, "")
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("push image %s multi-arch manifest failed", imageName))
		}
		logger.Log.Infof("Digest: %s Length: %d", digest, length)
	}

	return nil
}
