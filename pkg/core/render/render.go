package render

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strings"
	"text/template"
)

var (
	regDoublePat = regexp.MustCompile(`\{\{.*?\}\}`)
	regRep       = regexp.MustCompile(`[\{\} ]`)
)

func Render(input string, replaceVars map[string]string) (string, error) {
	res := input
	oldVars := regDoublePat.FindAllString(input, -1)
	for _, v := range oldVars {
		// delete the " ", "{" and "}"
		realVar := regRep.ReplaceAllString(v, "")
		if shoutVar, ok := replaceVars[realVar]; !ok || regDoublePat.MatchString(shoutVar) {
			return "", fmt.Errorf("render failed: [%s]", realVar)
		} else {
			res = strings.ReplaceAll(res, v, shoutVar)
		}
	}
	return res, nil
}

type KubeKeyTemplate struct {
	Err      error
	Template *template.Template
	Buf      *strings.Builder
}

func NewKebeKeyTemplate(template *template.Template) *KubeKeyTemplate {
	var buf strings.Builder
	return &KubeKeyTemplate{
		Template: template,
		Buf:      &buf,
	}
}

// todo: some bug in buffer
func (k *KubeKeyTemplate) PauseRender(variables map[string]interface{}) {

	if err := k.Template.Execute(k.Buf, variables); err != nil {
		k.Err = err
	}
}

func (k *KubeKeyTemplate) String() (string, error) {
	if k.Err != nil {
		return "", errors.Wrap(k.Err, "Failed to render template")
	}
	return k.Buf.String(), nil
}
