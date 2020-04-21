module github.com/kubekey

go 1.13

require (
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/pixiake/kubekey v0.0.0-20200421102531-1d5b364e8c3b // indirect
	github.com/pkg/sftp v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/tmc/scp v0.0.0-20170824174625-f7b48647feef // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
