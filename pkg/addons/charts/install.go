/*
Copyright 2020 The KubeSphere Authors.

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

package charts

import (
	"context"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"time"
)

func debug(format string, v ...interface{}) {
	if false {
		format = fmt.Sprintf("[debug] %s\n", format)
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

func InstallChart(mgr *manager.Manager, addon *kubekeyapiv1alpha1.Addon, kubeconfig string) error {
	actionConfig := new(action.Configuration)
	var settings = cli.New()
	helmDriver := os.Getenv("HELM_DRIVER")
	settings.KubeConfig = kubeconfig
	var namespace string
	if addon.Namespace != "" {
		namespace = addon.Namespace
	} else {
		namespace = "default"
	}

	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, helmDriver, debug); err != nil {
		mgr.Logger.Fatal(err)
	}

	valueOpts := &values.Options{}
	if len(addon.Sources.Chart.Values) != 0 {
		valueOpts.Values = addon.Sources.Chart.Values
		if mgr.InCluster && addon.Sources.Chart.Name == "ks-installer" {
			config, err := rest.InClusterConfig()
			if err != nil {
				return err
			}
			// creates the clientset
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}
			s, _ := clientset.CoreV1().Secrets("kubesphere-system").Get(context.TODO(), "kubesphere-secret", metav1.GetOptions{})
			valueOpts.Values = append(valueOpts.Values, fmt.Sprintf("authentication.jwtSecret=%s", string(s.Data["secret"])))
		}
	}
	if len(addon.Sources.Chart.ValuesFile) != 0 {
		valueOpts.ValueFiles = []string{addon.Sources.Chart.ValuesFile}
	}

	client := action.NewUpgrade(actionConfig)

	var chartName string
	if addon.Sources.Chart.Name != "" {
		if addon.Sources.Chart.Repo == "" && addon.Sources.Chart.Path != "" {
			fmt.Println(addon.Sources.Chart.Repo)
			chartName = filepath.Join(addon.Sources.Chart.Path, addon.Sources.Chart.Name)
		} else {
			chartName = addon.Sources.Chart.Name
		}
	} else {
		mgr.Logger.Fatalln("No chart name is specified")
	}

	args := []string{addon.Name, chartName}

	client.Install = true
	client.Namespace = namespace
	client.Timeout = 300 * time.Second
	client.Keyring = defaultKeyring()
	client.RepoURL = addon.Sources.Chart.Repo
	client.Version = addon.Sources.Chart.Version
	//client.Force = true

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	if client.Install {
		histClient := action.NewHistory(actionConfig)
		histClient.Max = 1
		if _, err := histClient.Run(addon.Name); err == driver.ErrReleaseNotFound {
			fmt.Printf("Release %q does not exist. Installing it now.\n", addon.Name)
			instClient := action.NewInstall(actionConfig)
			instClient.CreateNamespace = true
			instClient.Namespace = client.Namespace
			instClient.Timeout = client.Timeout
			instClient.Keyring = client.Keyring
			instClient.RepoURL = client.RepoURL
			instClient.Version = client.Version

			release, err := runInstall(mgr, args, instClient, valueOpts, settings)
			if err != nil {
				return err
			}
			printReleaseInfo(release)
			return nil
		} else if err != nil {
			return err
		}
	}

	chartPath, err := client.ChartPathOptions.LocateChart(args[1], settings)
	if err != nil {
		return err
	}

	vals, err := valueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return err
	}

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return err
		}
	}

	if ch.Metadata.Deprecated {
		mgr.Logger.Warningln("This chart is deprecated")
	}

	release, err1 := client.Run(args[0], ch, vals)
	if err1 != nil {
		return errors.Wrap(err1, "UPGRADE FAILED")
	}
	printReleaseInfo(release)
	return nil
}

func runInstall(mgr *manager.Manager, args []string, client *action.Install, valueOpts *values.Options, settings *cli.EnvSettings) (*release.Release, error) {
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}
	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		mgr.Logger.Warningln("This chart is deprecated")
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	return client.Run(chartRequested, vals)
}

func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func defaultKeyring() string {
	if v, ok := os.LookupEnv("GNUPGHOME"); ok {
		return filepath.Join(v, "pubring.gpg")
	}
	return filepath.Join(homedir.HomeDir(), ".gnupg", "pubring.gpg")
}

func printReleaseInfo(release *release.Release) {
	fmt.Printf("NAME: %s\n", release.Name)
	if !release.Info.LastDeployed.IsZero() {
		fmt.Printf("LAST DEPLOYED: %s\n", release.Info.LastDeployed.Format(time.ANSIC))
	}
	fmt.Printf("NAMESPACE: %s\n", release.Namespace)
	fmt.Printf("STATUS: %s\n", release.Info.Status.String())
	fmt.Printf("REVISION: %d\n", release.Version)
}
