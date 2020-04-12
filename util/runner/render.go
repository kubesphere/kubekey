package runner

import (
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(cmd string, variables map[string]interface{}) (string, error) {
	tpl, err := template.New("base").Parse(cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse script template")
	}

	var buf strings.Builder
	buf.WriteString(`set -xeu pipefail`)
	buf.WriteString("\n\n")
	buf.WriteString(`export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"`)
	buf.WriteString("\n\n")

	if err := tpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "failed to render script template")
	}

	return buf.String(), nil
}
