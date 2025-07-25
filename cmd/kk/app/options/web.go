package options

import (
	"fmt"
	"os"
	"path/filepath"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

// defaultPort defines the default port number for the web server
const (
	defaultPort = 80
)

// KubeKeyWebOptions contains configuration options for the KubeKey web server
type KubeKeyWebOptions struct {
	Port    int    // Port specifies the port number for the web server
	Workdir string // Workdir specifies the base directory for KubeKey

	SchemaPath string
	UIPath     string
}

// NewKubeKeyWebOptions creates and returns a new KubeKeyWebOptions instance with default values
func NewKubeKeyWebOptions() *KubeKeyWebOptions {
	o := &KubeKeyWebOptions{
		Port: defaultPort,
	}
	// Set the working directory to the current directory joined with "kubekey".
	wd, err := os.Getwd()
	if err != nil {
		klog.ErrorS(err, "get current dir error")
		o.Workdir = "/root/kubekey"
	} else {
		o.Workdir = filepath.Join(wd, "kubekey")
	}

	return o
}

// Flags returns a NamedFlagSets object containing command-line flags for configuring the web server
func (o *KubeKeyWebOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	wfs := fss.FlagSet("web flags")
	wfs.IntVar(&o.Port, "port", o.Port, fmt.Sprintf("the server port of kubekey web default is: %d", o.Port))
	wfs.StringVar(&o.Workdir, "workdir", o.Workdir, "the base Dir for kubekey. Default current dir. ")
	wfs.StringVar(&o.SchemaPath, "schema-path", o.SchemaPath, "the json schema dir path to render web ui.")
	wfs.StringVar(&o.UIPath, "ui-path", o.SchemaPath, "the web ui package path.")

	return fss
}
