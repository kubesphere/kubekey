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

package config

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func ParseClusterCfg(clusterCfgPath string, all bool, logger *log.Logger) (*kubekeyapi.Cluster, error) {

	if len(clusterCfgPath) == 0 {
		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return nil, errors.Wrap(err, "Failed to look up current directory")
		}
		cfgFile := fmt.Sprintf("%s/config.yaml", currentDir)
		if util.IsExist(cfgFile) {
			clusterCfg, err := ParseCfg(cfgFile)
			if err != nil {
				return nil, err
			}
			return clusterCfg, nil
		} else {
			user, _ := user.Current()
			if user.Name != "root" {
				return nil, errors.New(fmt.Sprintf("Current user is %s. Please use root!", user.Name))
			}
			clusterCfg := AllinoneCfg(user, all)
			return clusterCfg, nil
		}
	} else {
		clusterCfg, err := ParseCfg(clusterCfgPath)
		if err != nil {
			return nil, err
		}

		//output, _ := json.MarshalIndent(&clusterCfg, "", "  ")
		//fmt.Println(string(output))
		return clusterCfg, nil
	}
}

func ParseCfg(clusterCfgPath string) (*kubekeyapi.Cluster, error) {
	clusterCfg := kubekeyapi.Cluster{}
	fp, err := filepath.Abs(clusterCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to look up current directory")
	}
	content, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read the given cluster configuration file")
	}

	if err := yaml.Unmarshal(content, &clusterCfg); err != nil {
		return nil, errors.Wrap(err, "Unable to convert file to yaml")
	}
	return &clusterCfg, nil
}

func AllinoneCfg(user *user.User, all bool) *kubekeyapi.Cluster {
	allinoneCfg := kubekeyapi.Cluster{}
	if err := exec.Command("/bin/sh", "-c", "if [ ! -f \"$HOME/.ssh/id_rsa\" ]; then ssh-keygen -t rsa -P \"\" -f $HOME/.ssh/id_rsa && ls $HOME/.ssh;fi;").Run(); err != nil {
		log.Fatalf("Failed to generate public key: %v", err)
	}
	if output, err := exec.Command("/bin/sh", "-c", "echo \"\n$(cat $HOME/.ssh/id_rsa.pub)\" >> $HOME/.ssh/authorized_keys && awk ' !x[$0]++{print > \"'$HOME'/.ssh/authorized_keys\"}' $HOME/.ssh/authorized_keys").CombinedOutput(); err != nil {
		log.Fatalf("Failed to copy public key to authorized_keys: %v\n%s", err, string(output))
	}

	allinoneCfg.Spec.Hosts = append(allinoneCfg.Spec.Hosts, kubekeyapi.HostCfg{
		Name:            "ks-allinone",
		Address:         util.LocalIP(),
		InternalAddress: util.LocalIP(),
		Port:            "",
		User:            user.Name,
		Password:        "",
		PrivateKeyPath:  fmt.Sprintf("%s/.ssh/id_rsa", user.HomeDir),
	})

	allinoneCfg.Spec.RoleGroups = kubekeyapi.RoleGroups{
		Etcd:   []string{"ks-allinone"},
		Master: []string{"ks-allinone"},
		Worker: []string{"ks-allinone"},
	}

	if all {
		allinoneCfg.Spec.Storage = kubekeyapi.Storage{
			DefaultStorageClass: "local",
			LocalVolume:         kubekeyapi.LocalVolume{StorageClassName: "local"},
		}
		allinoneCfg.Spec.KubeSphere = kubekeyapi.KubeSphere{
			Console: kubekeyapi.Console{
				EnableMultiLogin: false,
				Port:             30880,
			},
			Common: kubekeyapi.Common{
				MysqlVolumeSize:    "20Gi",
				MinioVolumeSize:    "20Gi",
				EtcdVolumeSize:     "20Gi",
				OpenldapVolumeSize: "2Gi",
				RedisVolumSize:     "2Gi",
			},
			Openpitrix: kubekeyapi.Openpitrix{Enabled: false},
			Monitoring: kubekeyapi.Monitoring{
				PrometheusReplicas:      1,
				PrometheusMemoryRequest: "400Mi",
				PrometheusVolumeSize:    "20Gi",
				Grafana:                 kubekeyapi.Grafana{Enabled: false},
			},
			Logging: kubekeyapi.Logging{
				Enabled:                       false,
				ElasticsearchMasterReplicas:   1,
				ElasticsearchDataReplicas:     1,
				LogsidecarReplicas:            2,
				ElasticsearchMasterVolumeSize: "4Gi",
				ElasticsearchDataVolumeSize:   "20Gi",
				LogMaxAge:                     7,
				ElkPrefix:                     "logstash",
				Kibana:                        kubekeyapi.Kibana{Enabled: false},
			},
			Devops: kubekeyapi.Devops{
				Enabled:               false,
				JenkinsMemoryLim:      "2Gi",
				JenkinsMemoryReq:      "1500Mi",
				JenkinsVolumeSize:     "8Gi",
				JenkinsJavaOptsXms:    "512m",
				JenkinsJavaOptsXmx:    "512m",
				JenkinsJavaOptsMaxRAM: "2g",
				Sonarqube: kubekeyapi.Sonarqube{
					Enabled:              false,
					PostgresqlVolumeSize: "8Gi",
				},
			},
			Notification:  kubekeyapi.Notification{Enabled: false},
			Alerting:      kubekeyapi.Alerting{Enabled: false},
			ServiceMesh:   kubekeyapi.ServiceMesh{Enabled: false},
			MetricsServer: kubekeyapi.MetricsServer{Enabled: false},
		}

	}
	return &allinoneCfg
}
