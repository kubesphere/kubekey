package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/core/common"
	"github.com/kubesphere/kubekey/experiment/core/util"
	"github.com/kubesphere/kubekey/experiment/core/vars"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"text/template"
)

type Template struct {
	BaseAction
	TemplateName string
	Dst          string
	Data         map[string]interface{}
}

func (t *Template) Execute(vars vars.Vars) error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("template %s get pwd failed", t.TemplateName))
	}

	path := filepath.Join(pwd, common.ModuleTemplateDir, t.TemplateName)
	tmpl, err := template.New(t.TemplateName).ParseFiles(path)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("template %s parse failed", path))
	}

	templateStr, err := util.Render(tmpl, t.Data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("render template %s failed", t.TemplateName))
	}

	fileName := filepath.Join(t.Runtime.WorkDir, t.TemplateName)
	if err := util.WriteFile(fileName, []byte(templateStr)); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write file %s failed", fileName))
	}

	if err := t.Runtime.Runner.Scp(fileName, t.Dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fileName, t.Dst))
	}

	return nil
}
