package logger

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/common"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"time"
)

var Log *KubeKeyLog

func init() {
	Log = NewLogger()
}

type KubeKeyLog struct {
	logrus.FieldLogger
	RootEntry logrus.FieldLogger
}

func NewLogger() *KubeKeyLog {
	logger := logrus.New()

	formatter := &Formatter{
		HideKeys:               true,
		TimestampFormat:        "15:04:05 MST",
		NoColors:               true,
		ShowLevel:              logrus.FatalLevel,
		FieldsDisplayWithOrder: []string{common.Pipeline, common.Module, common.Task, common.Node},
	}

	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.InfoLevel)

	path := "./kubekey.log"
	writer, _ := rotatelogs.New(
		path+".%Y%m%d",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	logger.Hooks.Add(lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, formatter))

	return &KubeKeyLog{logger, logger}
}

func (k *KubeKeyLog) Message(node, str string) {
	Log.Infof("message: [%s]\n%s", node, str)
}

func (k *KubeKeyLog) Messagef(node, format string, args ...interface{}) {
	Log.Infof("message: [%s]\n%s", node, fmt.Sprintf(format, args...))
}
