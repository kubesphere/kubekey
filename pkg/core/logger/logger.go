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
	RootEntry logrus.FieldLogger
}

func NewLogger() *KubeKeyLog {
	logger := logrus.New()

	formatter := &nested.Formatter{
		HideKeys:        true,
		TimestampFormat: "15:04:05 MST",
		NoColors:        true,
		FieldsOrder:     []string{"Pipeline", "Module", "Task", "Node"},
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

	return &KubeKeyLog{logger, logger}
}

func (l *KubeKeyLog) Flush() {
	l.FieldLogger = l.RootEntry
}

func (l *KubeKeyLog) SetPipeline(pipeline string) {
	l.FieldLogger = l.WithFields(logrus.Fields{
		"Pipeline": pipeline,
	})
	l.RootEntry = l.FieldLogger
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
		"Node": node,
	})
}
