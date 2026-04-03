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

package image

import (
	"context"
	"io"
	"regexp"

	"github.com/cockroachdb/errors"
	"k8s.io/utils/ptr"

	"github.com/kubesphere/kubekey/v4/pkg/utils"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// Deprecated: Use the new configuration format (src/dest/manifests) instead.
// oldImagePullArgs contains parameters for pulling images from remote registries to local directories.
// This struct is maintained for backward compatibility with older configuration formats.
type oldImagePullArgs struct {
	imagesDir     string
	manifests     []string
	skipTLSVerify *bool
	platform      []string
	auths         []imageAuth
}

// Deprecated: Use the new configuration format (src/dest/manifests) instead.
// oldImagePushArgs contains parameters for pushing images from local directories to remote registries.
// This struct is maintained for backward compatibility with older configuration formats.
type oldImagePushArgs struct {
	imagesDir     string
	skipTLSVerify *bool
	srcPattern    *regexp.Regexp
	destTmpl      string
	auths         []imageAuth
}

// Deprecated: Use the new configuration format (src/dest/manifests) instead.
// oldImageCopyArgs contains parameters for copying images between local directories.
// This struct is maintained for backward compatibility with older configuration formats.
type oldImageCopyArgs struct {
	Platform []string               `json:"platform"`
	From     oldImageCopyTargetArgs `json:"from"`
	To       oldImageCopyTargetArgs `json:"to"`
}

// Deprecated: Use the new configuration format (src/dest/manifests) instead.
// oldImageCopyTargetArgs contains the source or destination configuration for image copy operations.
// This struct is maintained for backward compatibility with older configuration formats.
type oldImageCopyTargetArgs struct {
	Path      string `json:"path"`
	manifests []string
	Pattern   *regexp.Regexp
}

// transferPull parses deprecated "pull" configuration arguments and converts them to the new imageArgs format.
// Old format: pull images from remote registry to local directory
// New format: src = remote reference (oci://), dest = local directory (local://)
func transferPull(ctx context.Context, pullArgs any, vars map[string]any, logOutput io.Writer) (*imageArgs, error) {
	pull, ok := pullArgs.(map[string]any)
	if !ok {
		return nil, errors.New("\"pull\" should be map")
	}

	ipl := &oldImagePullArgs{}
	tpl := utils.GetTmpl(ctx)
	ipl.manifests, _ = variable.StringSliceVar(tpl, vars, pull, "manifests")
	ipl.auths = make([]imageAuth, 0)
	pullAuths := make([]imageAuth, 0)
	_ = variable.AnyVar(tpl, vars, pull, &pullAuths, "auths")
	ipl.auths = append(ipl.auths, pullAuths...)

	ipl.imagesDir, _ = variable.StringVar(tpl, vars, pull, "images_dir")
	ipl.skipTLSVerify, _ = variable.BoolVar(tpl, vars, pull, "skip_tls_verify")
	if ipl.skipTLSVerify == nil {
		ipl.skipTLSVerify = ptr.To(false)
	}
	ipl.platform, _ = variable.StringSliceVar(tpl, vars, pull, "platform")

	// Validate required fields
	if len(ipl.manifests) == 0 {
		return nil, errors.New("\"pull.manifests\" is required")
	}

	// Convert to new format
	ia := &imageArgs{
		manifests: ipl.manifests,
		platform:  ipl.platform,
		auths:     ipl.auths,
		src:       "oci://{{ .module.image.reference }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}",
		dest:      "local://" + ipl.imagesDir,
		logOutput: logOutput,
	}

	return ia, nil
}

// transferPush parses deprecated "push" configuration arguments and converts them to the new imageArgs format.
// Old format: push images from local directory to remote registry
// New format: src = local directory (local://), dest = remote reference (oci://)
func transferPush(ctx context.Context, pushArgs any, vars map[string]any, logOutput io.Writer) (*imageArgs, error) {
	push, ok := pushArgs.(map[string]any)
	if !ok {
		return nil, errors.New("\"push\" should be map")
	}

	ips := &oldImagePushArgs{}
	ips.auths = make([]imageAuth, 0)
	pushAuths := make([]imageAuth, 0)
	tpl := utils.GetTmpl(ctx)
	_ = variable.AnyVar(tpl, vars, push, &pushAuths, "auths")
	ips.auths = append(ips.auths, pushAuths...)

	ips.imagesDir, _ = variable.StringVar(tpl, vars, push, "images_dir")
	srcPattern, _ := variable.StringVar(tpl, vars, push, "src_pattern")
	destTmpl, _ := variable.PrintVar(push, "dest")
	ips.skipTLSVerify, _ = variable.BoolVar(tpl, vars, push, "skip_tls_verify")
	if ips.skipTLSVerify == nil {
		ips.skipTLSVerify = ptr.To(false)
	}

	// Validate required fields
	if srcPattern != "" {
		pattern, err := regexp.Compile(srcPattern)
		if err != nil {
			return nil, errors.Wrap(err, "\"push.src\" should be a valid regular expression. ")
		}
		ips.srcPattern = pattern
	}
	if destStr, ok := destTmpl.(string); !ok {
		return nil, errors.New("\"push.dest\" must be a string")
	} else if destStr == "" {
		return nil, errors.New("\"push.dest\" should not be empty")
	} else {
		ips.destTmpl = destStr
	}

	// Convert to new format
	ia := &imageArgs{
		src:       "local://" + ips.imagesDir,
		dest:      "oci://" + ips.destTmpl,
		pattern:   ips.srcPattern,
		auths:     ips.auths,
		logOutput: logOutput,
	}
	if ips.skipTLSVerify != nil {
		for i := range ia.auths {
			ia.auths[i].SkipTLSVerify = ips.skipTLSVerify
		}
	}

	return ia, nil
}

// transferCopy parses deprecated "copy" configuration arguments and converts them to the new imageArgs format.
// Old format: copy images from local directory to local directory
// New format: src = local directory (local://), dest = local directory (local://)
func transferCopy(ctx context.Context, copyArgs any, vars map[string]any, logOutput io.Writer) (*imageArgs, error) {
	cp, ok := copyArgs.(map[string]any)
	if !ok {
		return nil, errors.New("\"copy\" should be map")
	}

	cps := &oldImageCopyArgs{}

	tpl := utils.GetTmpl(ctx)
	cps.Platform, _ = variable.StringSliceVar(tpl, vars, cp, "platform")

	cps.From.manifests, _ = variable.StringSliceVar(tpl, vars, cp, "from", "manifests")

	cps.From.Path, _ = variable.StringVar(tpl, vars, cp, "from", "path")

	toPath, _ := variable.PrintVar(cp, "to", "path")
	if destStr, ok := toPath.(string); !ok {
		return nil, errors.New("\"copy.to.path\" must be a string")
	} else if destStr == "" {
		return nil, errors.New("\"copy.to.path\" should not be empty")
	} else {
		cps.To.Path = destStr
	}
	srcPattern, _ := variable.StringVar(tpl, vars, cp, "to", "src_pattern")
	if srcPattern != "" {
		pattern, err := regexp.Compile(srcPattern)
		if err != nil {
			return nil, errors.Wrap(err, "\"to.pattern\" should be a valid regular expression. ")
		}
		cps.From.Pattern = pattern
	}

	// Convert to new format
	ia := &imageArgs{
		src:       "local://" + cps.From.Path,
		dest:      "local://" + cps.To.Path,
		platform:  cps.Platform,
		manifests: cps.From.manifests,
		pattern:   cps.From.Pattern,
		logOutput: logOutput,
	}

	return ia, nil
}
