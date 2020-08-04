module github.com/kubesphere/kubekey

go 1.14

require (
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/imdario/mergo v0.3.10 // indirect
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
	golang.org/x/tools v0.0.0-20200403190813-44a64ad78b9b
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20200724153422-f32512634ab7 // indirect
	sigs.k8s.io/controller-runtime v0.6.1
)

replace k8s.io/client-go => k8s.io/client-go v0.18.0
