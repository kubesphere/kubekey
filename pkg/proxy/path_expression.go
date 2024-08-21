/*
Copyright 2024 The KubeSphere Authors.

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

package proxy

// Copyright 2013 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// PathExpression holds a compiled path expression (RegExp) needed to match against
// Http request paths and to extract path parameter values.
type pathExpression struct {
	LiteralCount int      // the number of literal characters (means those not resulting from template variable substitution)
	VarNames     []string // the names of parameters (enclosed by {}) in the path
	VarCount     int      // the number of named parameters (enclosed by {}) in the path
	Matcher      *regexp.Regexp
	Source       string // Path as defined by the RouteBuilder
	tokens       []string
}

// NewPathExpression creates a PathExpression from the input URL path.
// Returns an error if the path is invalid.
func newPathExpression(path string) (*pathExpression, error) {
	expression, literalCount, varNames, varCount, tokens := templateToRegularExpression(path)
	compiled, err := regexp.Compile(expression)
	if err != nil {
		return nil, err
	}

	return &pathExpression{literalCount, varNames, varCount, compiled, expression, tokens}, nil
}

// http://jsr311.java.net/nonav/releases/1.1/spec/spec3.html#x3-370003.7.3
func templateToRegularExpression(template string) (string, int, []string, int, []string) {
	var (
		literalCount int
		varNames     []string
		varCount     int
	)
	var buffer bytes.Buffer
	buffer.WriteString("^")
	tokens := tokenizePath(template)
	for _, each := range tokens {
		if each == "" {
			continue
		}
		buffer.WriteString("/")
		if strings.HasPrefix(each, "{") {
			// check for regular expression in variable
			colon := strings.Index(each, ":")
			var varName string
			if colon != -1 {
				// extract expression
				varName = strings.TrimSpace(each[1:colon])
				paramExpr := strings.TrimSpace(each[colon+1 : len(each)-1])
				if paramExpr == "*" { // special case
					buffer.WriteString("(.*)")
				} else {
					buffer.WriteString(fmt.Sprintf("(%s)", paramExpr)) // between colon and closing moustache
				}
			} else {
				// plain var
				varName = strings.TrimSpace(each[1 : len(each)-1])
				buffer.WriteString("([^/]+?)")
			}
			varNames = append(varNames, varName)
			varCount++
		} else {
			literalCount += len(each)
			encoded := each
			buffer.WriteString(regexp.QuoteMeta(encoded))
		}
	}

	return strings.TrimRight(buffer.String(), "/") + "(/.*)?$", literalCount, varNames, varCount, tokens
}

// Tokenize an URL path using the slash separator ; the result does not have empty tokens
func tokenizePath(path string) []string {
	if path == "/" {
		return nil
	}
	// 3.9.0
	return strings.Split(strings.Trim(path, "/"), "/")
}
