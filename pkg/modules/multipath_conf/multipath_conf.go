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
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const defaultPath = "/etc/multipath.conf"

type item struct {
	Path     string
	Backup   bool
	Reload   bool
	DevNodes []string
}

func ModuleMultipathConf(ctx context.Context, opts internal.ExecOptions) (string, string, error) {
	args := variable.Extension2Variables(opts.Args)
	cfg, err := parseItem(args)
	if err != nil {
		return internal.StdoutFailed, "invalid multipath configuration", err
	}

	conn, err := opts.GetConnector(ctx)
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetConnector, err
	}
	defer conn.Close(ctx)

	stdout, stderr, err := conn.ExecuteCommand(ctx, buildScript(cfg))
	if err != nil {
		return string(stdout), string(stderr), err
	}

	return string(stdout), string(stderr), nil
}

func parseItem(args map[string]any) (item, error) {
	cfg := item{
		Path:     defaultPath,
		Backup:   true,
		Reload:   true,
		DevNodes: []string{"^sd[a-z]", "^vd[a-z]", "^xvd[a-z]", "^nvme[0-9]n[0-9]"},
	}

	if path, err := variable.StringVar(nil, args, "path"); err == nil && strings.TrimSpace(path) != "" {
		cfg.Path = strings.TrimSpace(path)
	}
	if backup, err := variable.BoolVar(nil, args, "backup"); err == nil && backup != nil {
		cfg.Backup = *backup
	}
	if reload, err := variable.BoolVar(nil, args, "reload"); err == nil && reload != nil {
		cfg.Reload = *reload
	}
	if devNodes, err := variable.StringSliceVar(nil, args, "devnodes"); err == nil && len(devNodes) > 0 {
		cfg.DevNodes = nonEmptyTrimmed(devNodes)
	}

	if cfg.Path == "" {
		return item{}, errors.New("path is required")
	}
	if len(cfg.DevNodes) == 0 {
		return item{}, errors.New("devnodes must contain at least one rule")
	}

	return cfg, nil
}

func nonEmptyTrimmed(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func buildScript(cfg item) string {
	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail

PATH_VALUE=%q
BACKUP=%t
RELOAD=%t
DEVNODES=(%s)

ensure_blacklist_block() {
  if [ ! -f "$PATH_VALUE" ]; then
    cat > "$PATH_VALUE" <<'EOF'
blacklist {
}
EOF
    return 0
  fi

  if ! grep -Eq '^blacklist[[:space:]]*\{' "$PATH_VALUE"; then
    cat >> "$PATH_VALUE" <<'EOF'

blacklist {
}
EOF
  fi
}

ensure_devnode_rule() {
  local rule="$1"
  if grep -Eq "^[[:space:]]*devnode[[:space:]]+\"${rule//\\/\\\\}\"" "$PATH_VALUE"; then
    return 0
  fi

  awk -v rule="$rule" '
    BEGIN { inserted = 0; depth = 0; in_blacklist = 0 }
    /^[[:space:]]*blacklist[[:space:]]*\{/ { in_blacklist = 1 }
    {
      if (in_blacklist) {
        depth += gsub(/\{/, "{")
        depth -= gsub(/\}/, "}")
        if (depth == 0 && inserted == 0) {
          print "    devnode \"" rule "\""
          inserted = 1
        }
      }
      print
    }
  ' "$PATH_VALUE" > "$PATH_VALUE.tmp"
  install -m 0644 "$PATH_VALUE.tmp" "$PATH_VALUE"
  rm -f "$PATH_VALUE.tmp"
}

mkdir -p "$(dirname "$PATH_VALUE")"
if [ "$BACKUP" = true ] && [ -f "$PATH_VALUE" ]; then
  cp -p "$PATH_VALUE" "$PATH_VALUE.$(date -u +%%Y%%m%%d_%%H%%M%%S).bak"
fi

ensure_blacklist_block
for devnode in "${DEVNODES[@]}"; do
  ensure_devnode_rule "$devnode"
done

if command -v multipath >/dev/null 2>&1; then
  multipath -t >/dev/null
fi

if [ "$RELOAD" = true ]; then
  if command -v systemctl >/dev/null 2>&1 && systemctl list-unit-files | grep -q '^multipathd\.service'; then
    systemctl reload multipathd || systemctl restart multipathd
  elif command -v service >/dev/null 2>&1 && service multipathd status >/dev/null 2>&1; then
    service multipathd reload || service multipathd restart
  else
    echo "multipathd service not found, skip reload"
  fi
fi
`, cfg.Path, cfg.Backup, cfg.Reload, quoteBashWords(cfg.DevNodes))
}

func quoteBashWords(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, " ")
}
