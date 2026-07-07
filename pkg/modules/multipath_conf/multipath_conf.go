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
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
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

	stdout, err := applyMultipathConf(ctx, conn, cfg)
	if err != nil {
		return internal.StdoutFailed, stdout, err
	}

	return stdout, "", nil
}

func applyMultipathConf(ctx context.Context, conn connector.Connector, cfg item) (string, error) {
	original, exists, err := readRemoteFile(ctx, conn, cfg.Path)
	if err != nil {
		return "", err
	}

	updated, changed, err := updateMultipathConfig(string(original), cfg.DevNodes)
	if err != nil {
		return "", err
	}
	if !changed {
		klog.V(4).InfoS("multipath config already contains requested devnode rules", "path", cfg.Path)
	} else if cfg.Backup && exists {
		backupPath := fmt.Sprintf("%s.%s.bak", cfg.Path, time.Now().UTC().Format("20060102_150405"))
		if err := connector.PutData(ctx, original, backupPath, _const.PermFilePublic, conn); err != nil {
			return "", errors.Wrapf(err, "failed to backup multipath config to %s", backupPath)
		}
		klog.V(4).InfoS("backed up multipath config", "path", cfg.Path, "backup", backupPath)
	}

	if changed {
		if err := connector.PutData(ctx, []byte(updated), cfg.Path, _const.PermFilePublic, conn); err != nil {
			return "", errors.Wrapf(err, "failed to write multipath config %s", cfg.Path)
		}
	}

	if err := validateMultipathConfig(ctx, conn); err != nil {
		return "", err
	}
	if cfg.Reload {
		if err := reloadMultipathService(ctx, conn); err != nil {
			return "", err
		}
	}

	return "multipath config updated", nil
}

func readRemoteFile(ctx context.Context, conn connector.Connector, path string) ([]byte, bool, error) {
	var buf bytes.Buffer
	if err := conn.FetchFile(ctx, path, &buf); err != nil {
		if isRemoteNotExist(err) {
			return nil, false, nil
		}
		return nil, false, errors.Wrapf(err, "failed to read multipath config %s", path)
	}
	return buf.Bytes(), true, nil
}

func isRemoteNotExist(err error) bool {
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "file does not exist")
}

func validateMultipathConfig(ctx context.Context, conn connector.Connector) error {
	if _, _, err := conn.ExecuteCommand(ctx, "command -v multipath >/dev/null 2>&1"); err != nil {
		klog.V(4).InfoS("multipath command not found, skip config validation")
		return nil
	}

	_, stderr, err := conn.ExecuteCommand(ctx, "multipath -t")
	if err != nil {
		if len(bytes.TrimSpace(stderr)) > 0 {
			return errors.Wrapf(err, "multipath config validation failed: %s", strings.TrimSpace(string(stderr)))
		}
		return errors.Wrap(err, "multipath config validation failed")
	}
	return nil
}

func reloadMultipathService(ctx context.Context, conn connector.Connector) error {
	commands := []string{
		"systemctl reload multipathd",
		"systemctl restart multipathd",
		"service multipathd reload",
		"service multipathd restart",
	}
	for _, command := range commands {
		if _, _, err := conn.ExecuteCommand(ctx, command); err == nil {
			return nil
		}
	}
	klog.V(4).InfoS("multipathd service not found, skip reload")
	return nil
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
