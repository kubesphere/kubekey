package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"path/filepath"
	"text/template"
)

type Template struct {
	BaseAction
	Template *template.Template
	Dst      string
	Data     util.Data
}

func (t *Template) Execute(runtime connector.Runtime) error {
	templateStr, err := util.Render(t.Template, t.Data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("render template %s failed", t.Template.Name()))
	}

	fileName := filepath.Join(runtime.GetHostWorkDir(), t.Template.Name())
	if err := util.WriteFile(fileName, []byte(templateStr)); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write file %s failed", fileName))
	}

	if err := runtime.GetRunner().SudoScp(fileName, t.Dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fileName, t.Dst))
	}

	return nil
}
