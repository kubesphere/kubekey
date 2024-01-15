/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tmpl

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v6"
	"k8s.io/apimachinery/pkg/util/version"
)

func init() {
	pongo2.RegisterFilter("defined", filterDefined)
	pongo2.RegisterFilter("version", filterVersion)
	pongo2.RegisterFilter("pow", filterPow)
	pongo2.RegisterFilter("match", filterMatch)
	pongo2.RegisterFilter("basename", filterBasename)
}

func filterDefined(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.IsNil() {
		return pongo2.AsValue(false), nil
	}
	return pongo2.AsValue(true), nil
}

func filterVersion(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	inVersion, err := version.ParseGeneric(in.String())
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "filter:version in",
			OrigError: err,
		}
	}
	paramString := param.String()
	customChoices := strings.Split(paramString, ",")
	if len(customChoices) != 2 {
		return nil, &pongo2.Error{
			Sender:    "filter:version",
			OrigError: fmt.Errorf("'version'-filter need 2 arguments(as: verion:'xxx,xxx') but got'%s'", paramString),
		}
	}
	ci, err := inVersion.Compare(customChoices[1])
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "filter:version",
			OrigError: fmt.Errorf("converter second param error: %v", err),
		}
	}
	switch customChoices[0] {
	case ">":
		return pongo2.AsValue(ci == 1), nil
	case "=":
		return pongo2.AsValue(ci == 0), nil
	case "<":
		return pongo2.AsValue(ci == -1), nil
	case ">=":
		return pongo2.AsValue(ci >= 0), nil
	case "<=":
		return pongo2.AsValue(ci <= 0), nil
	default:
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "filter:version",
			OrigError: fmt.Errorf("converter first param error: %v", err),
		}
	}
}

func filterPow(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(math.Pow(in.Float(), param.Float())), nil
}

func filterMatch(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	match, err := regexp.Match(param.String(), []byte(in.String()))
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{Sender: "filter:match", OrigError: err}
	}
	return pongo2.AsValue(match), nil
}

func filterBasename(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(filepath.Base(in.String())), nil
}
