//go:build builtin
// +build builtin

package tmpl

import (
	"text/template"

	"github.com/cockroachdb/errors"

	"github.com/kubesphere/kubekey/v4/builtin/core"
)

func loadBuiltinIncludeTemplates(tl *template.Template) (err error) {
	if _, err = tl.ParseFS(core.BuiltinPlaybook, "tpls/*.tpl"); err != nil {
		err = errors.Wrapf(err, "failed to parse builtin template %q", "tpls/*.tpl")
	}
	return
}
