package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

// K3sServiceEnv defines the template of kubelet's Env for the kubelet's systemd service.
var K3sServiceEnv = template.Must(template.New("k3s.service.env").Parse(
	dedent.Dedent(`# Note: This dropin only works with k3s
{{ if .IsMaster }}
K3S_DATASTORE_ENDPOINT={{ .DataStoreEndPoint }}
K3S_DATASTORE_CAFILE={{ .DataStoreCaFile }}
K3S_DATASTORE_CERTFILE={{ .DataStoreCertFile }}
K3S_DATASTORE_KEYFILE={{ .DataStoreKeyFile }}
K3S_KUBECONFIG_MODE=644
{{ end }}
{{ if .Token }}
K3S_TOKEN={{ .Token }}
{{ end }}

    `)))
