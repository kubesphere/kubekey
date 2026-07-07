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

package multipath_conf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
)

var blacklistBlockPattern = regexp.MustCompile(`(?m)^[ \t]*blacklist[ \t]*\{`)

func updateMultipathConfig(content string, devnodes []string) (string, bool, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	changed := false

	if strings.TrimSpace(content) == "" {
		content = "blacklist {\n}\n"
		changed = true
	}

	if !blacklistBlockPattern.MatchString(content) {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\nblacklist {\n}\n"
		changed = true
	}

	for _, rule := range devnodes {
		if hasDevnodeRule(content, rule) {
			continue
		}
		updated, err := insertDevnodeRule(content, rule)
		if err != nil {
			return "", false, err
		}
		content = updated
		changed = true
	}

	return content, changed, nil
}

func hasDevnodeRule(content, rule string) bool {
	pattern := regexp.MustCompile(`(?m)^[ \t]*devnode[ \t]+"` + regexp.QuoteMeta(rule) + `"[ \t]*$`)
	return pattern.MatchString(content)
}

func insertDevnodeRule(content, rule string) (string, error) {
	match := blacklistBlockPattern.FindStringIndex(content)
	if match == nil {
		return "", errors.New("blacklist block not found")
	}

	start := strings.Index(content[match[0]:], "{")
	if start == -1 {
		return "", errors.New("blacklist opening brace not found")
	}
	start += match[0]

	depth := 0
	insertAt := -1
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				insertAt = i
			}
		}
		if insertAt >= 0 {
			break
		}
	}
	if insertAt < 0 {
		return "", errors.New("blacklist closing brace not found")
	}

	line := fmt.Sprintf("    devnode %q\n", rule)
	return content[:insertAt] + line + content[insertAt:], nil
}
