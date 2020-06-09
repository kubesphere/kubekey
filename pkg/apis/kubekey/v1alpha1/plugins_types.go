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

package v1alpha1

type Storage struct {
	DefaultStorageClass string      `yaml:"defaultStorageClass" json:"defaultStorageClass,omitempty"`
	LocalVolume         LocalVolume `yaml:"localVolume" json:"locaVolume,omitempty"`
	NfsClient           NfsClient   `yaml:"nfsClient" json:"nfs,omitempty"`
	GlusterFS           GlusterFS   `yaml:"glusterfs" json:"glusterfs,omitempty"`
	CephRBD             CephRBD     `yaml:"rbd" json:"rbd,omitempty"`
	NeonsanCSI          NeonsanCSI  `yaml:"neonsancsi" json:"neonsancsi"`
}

type LocalVolume struct {
	Enabled          bool   `json:"enabled,omitempty"`
	IsDefaultClass   bool   `json:"isDefaultClass,omitempty"`
	StorageClassName string `yaml:"storageClassName" json:"storageClassName,omitempty"`
}

type NfsClient struct {
	Enabled            bool   `json:"enabled,omitempty"`
	IsDefaultClass     bool   `json:"isDefaultClass,omitempty"`
	StorageClassName   string `yaml:"storageClassName" json:"storageClassName,omitempty"`
	NfsServer          string `yaml:"nfsServer" json:"nfsServer,omitempty"`
	NfsPath            string `yaml:"nfsPath" json:"nfsPath,omitempty"`
	NfsVrs3Enabled     bool   `yaml:"nfsVrs3Enabled" json:"nfsVrs3Enabled,omitempty"`
	NfsArchiveOnDelete bool   `yaml:"nfsArchiveOnDelete" json:"nfsArchiveOnDelete,omitempty"`
}

type GlusterFS struct {
	Enabled          bool   `json:"enabled,omitempty"`
	IsDefaultClass   bool   `json:"isDefaultClass,omitempty"`
	StorageClassName string `yaml:"storageClassName" json:"storageClassName,omitempty"`
	RestAuthEnabled  bool   `yaml:"restAuthEnabled" json:"restAuthEnabled,omitempty"`
	RestUrl          string `yaml:"restUrl" json:"restUrl,omitempty"`
	ClusterID        string `yaml:"clusterID" json:"clusterID,omitempty"`
	RestUser         string `yaml:"restUser" json:"restUser,omitempty"`
	SecretName       string `yaml:"secretName" json:"secretName,omitempty"`
	GidMin           int    `yaml:"gidMin" json:"gidMin,omitempty"`
	GidMax           int    `yaml:"gidMax" json:"gidMax,omitempty"`
	VolumeType       string `yaml:"volumeType" json:"volumeType,omitempty"`
	JwtAdminKey      string `yaml:"jwtAdminKey" json:"jwtAdminKey,omitempty"`
}

type CephRBD struct {
	Enabled          bool     `json:"enabled,omitempty"`
	IsDefaultClass   bool     `json:"isDefaultClass,omitempty"`
	StorageClassName string   `yaml:"storageClassName" json:"storageClassName,omitempty"`
	Monitors         []string `yaml:"monitors" json:"monitors,omitempty"`
	AdminID          string   `yaml:"adminID" json:"adminID,omitempty"`
	AdminSecret      string   `yaml:"adminSecret" json:"adminSecret,omitempty"`
	UserID           string   `yaml:"userID" json:"userID,omitempty"`
	UserSecret       string   `yaml:"userSecret" json:"userSecret,omitempty"`
	Pool             string   `yaml:"pool" json:"pool,omitempty"`
	FsType           string   `yaml:"fsType" json:"fsType,omitempty"`
	ImageFormat      int      `yaml:"imageFormat" json:"imageFormat,omitempty"`
	ImageFeatures    string   `yaml:"imageFeatures" json:"imageFeatures,omitempty"`
}

type NeonsanCSI struct {
	Enable  bool   `json:"enable,omitempty"`
	Pool    string `yaml:"pool" json:"pool"`
	Replica int    `yaml:"replica" json:"replica"`
	FsType  string `yaml:"fsType" json:"fsType"`
}
