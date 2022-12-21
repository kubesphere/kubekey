/*
 Copyright 2021 The KubeSphere Authors.

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

package logger

import (
	"fmt"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/kubesphere/kubekey/v2/pkg/core/common"
)

var Log *KubeKeyLog

type KubeKeyLog struct {
	logrus.FieldLogger
	OutputPath string
	Verbose    bool
}

func NewLogger(outputPath string, verbose bool) *KubeKeyLog {
	logger := logrus.New()

	formatter := &Formatter{
		HideKeys:               true,
		TimestampFormat:        "15:04:05 MST",
		NoColors:               true,
		ShowLevel:              logrus.WarnLevel,
		FieldsDisplayWithOrder: []string{common.Pipeline, common.Module, common.Task, common.Node},
	}
	logger.SetFormatter(formatter)

	path := filepath.Join(outputPath, "./kubekey.log")
	writer, _ := rotatelogs.New(
		path+".%Y%m%d",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	logWriters := lfshook.WriterMap{
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		logWriters[logrus.DebugLevel] = writer
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Hooks.Add(lfshook.NewHook(logWriters, formatter))
	return &KubeKeyLog{logger, outputPath, verbose}
}

func (k *KubeKeyLog) Message(node, str string) {
	Log.Infof("message: [%s]\n%s", node, str)
}

func (k *KubeKeyLog) Messagef(node, format string, args ...interface{}) {
	Log.Infof("message: [%s]\n%s", node, fmt.Sprintf(format, args...))
}
