package util

import (
	log "github.com/sirupsen/logrus"
)

const (
	VERSION = "KubeOcean Version v0.0.1\nKubernetes Version v1.17.4"
)

func InitLogger(verbose bool) *log.Logger {
	logger := log.New()
	logger.Formatter = &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 MST",
	}

	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	return logger
}
