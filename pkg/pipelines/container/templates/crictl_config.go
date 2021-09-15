package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

var CrictlConfig = template.Must(template.New("crictl.yaml").Parse(
	dedent.Dedent(`runtime-endpoint: {{ .Endpoint }}
image-endpoint: {{ .Endpoint }}
timeout: 5
debug: false
pull-image-on-create: false
    `)))
