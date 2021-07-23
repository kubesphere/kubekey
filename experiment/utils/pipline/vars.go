package pipline

import "text/template"

type Vars struct {
	configurations []map[string]interface{}
	hosts          []map[string]interface{}
	hostsInfo      []map[string]interface{}
}

func Render(tmpl template.Template, vars Vars) (string, error){
	return "", nil
}