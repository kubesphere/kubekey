package util

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"text/template"
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

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

type Data map[string]interface{}

// Render text template with given `variables` Render-context
func Render(tmpl *template.Template, variables map[string]interface{}) (string, error) {

	var buf strings.Builder
	//buf.WriteString(`set -xeu pipefail`)
	//buf.WriteString("\n\n")
	//buf.WriteString(`export "PATH=$PATH:/sbin:/usr/local/bin:/opt/bin"`)
	//buf.WriteString("\n\n")

	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "failed to render cmd or script template")
	}
	return buf.String(), nil
}
