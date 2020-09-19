module github.com/kubesphere/kubekey

go 1.14

require (
	github.com/dominodatalab/os-release v0.0.0-20190522011736-bcdb4a3e3c2f
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mitchellh/mapstructure v1.3.3
	github.com/modood/table v0.0.0-20200225102042-88de94bb9876
	github.com/operator-framework/operator-sdk v0.19.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/cli-runtime v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.18.6
	sigs.k8s.io/controller-runtime v0.6.0
)

replace k8s.io/client-go => k8s.io/client-go v0.18.6
