package kubesphere

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere/templates"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"text/template"
)

type KsInstaller struct {
	Version                      string
	CRDTemplate                  *template.Template
	ClusterConfigurationTemplate *template.Template
	K8sSupportVersions           []string
	UpgradeSupportVersions       []string
}

func (k *KsInstaller) CCToString() string {
	str, err := util.Render(k.ClusterConfigurationTemplate, util.Data{
		"Tag": k.Version,
	})
	if err != nil {
		os.Exit(0)
	}
	return str
}

func (k *KsInstaller) K8sSupport(version string) bool {
	k8sVersion := versionutil.MustParseSemantic(version)
	for i := range k.K8sSupportVersions {
		if k.K8sSupportVersions[i] == fmt.Sprintf("v%v.%v", k8sVersion.Major(), k8sVersion.Minor()) {
			return true
		}
	}
	return false
}

func (k *KsInstaller) UpgradeSupport(version string) bool {
	for i := range k.UpgradeSupportVersions {
		if k.UpgradeSupportVersions[i] == version {
			return true
		}
	}
	return false
}

var KsV211 = &KsInstaller{
	Version:                      V211.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V211,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
	},
	UpgradeSupportVersions: []string{
		V211.String(),
	},
}

var KsV300 = &KsInstaller{
	Version:                      V300.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V300,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
	},
	UpgradeSupportVersions: []string{
		V211.String(),
	},
}

var KsV310 = &KsInstaller{
	Version:                      V310.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V310,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
		"v1.19",
		"v1.20",
	},
	UpgradeSupportVersions: []string{
		V300.String(),
	},
}

var KsV311 = &KsInstaller{
	Version:                      V311.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V311,
	K8sSupportVersions: []string{
		"v1.15",
		"v1.16",
		"v1.17",
		"v1.18",
		"v1.19",
		"v1.20",
	},
	UpgradeSupportVersions: []string{
		V300.String(),
		V310.String(),
	},
}

var KsV320 = &KsInstaller{
	Version:                      V320.String(),
	CRDTemplate:                  templates.KsInstaller,
	ClusterConfigurationTemplate: templates.V320,
	K8sSupportVersions: []string{
		"v1.19",
		"v1.20",
		"v1.21",
		"v1.22",
	},
	UpgradeSupportVersions: []string{
		V310.String(),
		V311.String(),
	},
}
