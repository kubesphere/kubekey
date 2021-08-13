package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
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
}

func NewLogger() *KubeKeyLog {
	logger := logrus.New()

	formatter := &nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "15:04:05 MST",
		NoColors:        true,
	}

	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

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

	return &KubeKeyLog{logger}
}

func (l *KubeKeyLog) SetModule(module string) {
	l.FieldLogger = l.WithFields(logrus.Fields{
		"Module": module,
	})
}

func (l *KubeKeyLog) SetTask(task string) {
	l.FieldLogger = l.WithFields(logrus.Fields{
		"Task": task,
	})
}

func (l *KubeKeyLog) SetNode(node string) {
	l.FieldLogger = l.WithFields(logrus.Fields{
		"node": node,
	})
}
