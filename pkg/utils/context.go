package utils

import (
	"context"
	"text/template"
)

type tplKey struct{}

var TplKey = &tplKey{}

func GetTmpl(ctx context.Context) *template.Template {
	return ctx.Value(TplKey).(*template.Template)
}
