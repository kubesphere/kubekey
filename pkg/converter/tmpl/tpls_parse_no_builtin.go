//go:build !builtin
// +build !builtin

package tmpl

import (
	"text/template"
)

func loadBuiltinIncludeTemplates(tl *template.Template) error {
	return nil
}
