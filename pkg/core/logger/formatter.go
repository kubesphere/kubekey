package logger

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Formatter struct {
	// TimestampFormat - default: time.StampMilli = "Jan _2 15:04:05.000"
	TimestampFormat string
	// NoColors - disable colors
	NoColors bool
	// ShowLevel - when the level < this field, it won't be show. default: TRACE
	ShowLevel logrus.Level
	// ShowFullLevel - show a full level [WARNING] instead of [WARN]
	ShowFullLevel bool
	// NoUppercaseLevel - no upper case for level value
	NoUppercaseLevel bool
	// HideKeys - show [fieldValue] instead of [fieldKey:fieldValue]
	HideKeys bool
	// FieldsDisplayWithOrder - default: all fields display and sorted alphabetically
	FieldsDisplayWithOrder []string
	// CallerFirst - print caller info first
	CallerFirst bool
	// CustomCallerFormatter - set custom formatter for caller info
	CustomCallerFormatter func(*runtime.Frame) string
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getColorByLevel(entry.Level)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampMilli
	}

	// output buffer
	b := &bytes.Buffer{}

	// write time
	b.WriteString(entry.Time.Format(timestampFormat))

	if f.CallerFirst {
		f.writeCaller(b, entry)
	}

	if !f.NoColors {
		fmt.Fprintf(b, "\x1b[%dm", levelColor)
	}

	level := entry.Level
	if f.ShowLevel >= level {
		var levelStr string
		if f.NoUppercaseLevel {
			levelStr = entry.Level.String()
		} else {
			levelStr = strings.ToUpper(entry.Level.String())
		}

		b.WriteString(" [")
		if f.ShowFullLevel {
			b.WriteString(levelStr)
		} else {
			b.WriteString(levelStr[:4])
		}
		b.WriteString("]")
	}

	b.WriteString(" ")

	// write fields
	if f.FieldsDisplayWithOrder == nil {
		f.writeFields(b, entry)
	} else {
		f.writeOrderedFields(b, entry)
	}

	b.WriteString(entry.Message)

	if !f.CallerFirst {
		f.writeCaller(b, entry)
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return colorGray
	case logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return colorRed
	default:
		return colorBlue
	}
}

func (f *Formatter) writeFields(b *bytes.Buffer, entry *logrus.Entry) {
	if len(entry.Data) != 0 {
		fields := make([]string, 0, len(entry.Data))
		for field := range entry.Data {
			fields = append(fields, field)
		}

		sort.Strings(fields)

		b.WriteString("[")
		for i, field := range fields {
			f.writeField(b, entry, field, i)
		}
		b.WriteString("]")
	}
}

func (f *Formatter) writeOrderedFields(b *bytes.Buffer, entry *logrus.Entry) {
	if len(entry.Data) != 0 {
		b.WriteString("[")
		length := len(entry.Data)
		foundFieldsMap := map[string]bool{}
		for i, field := range f.FieldsDisplayWithOrder {
			if _, ok := entry.Data[field]; ok {
				foundFieldsMap[field] = true
				length--
				f.writeField(b, entry, field, i)
			}
		}

		if length > 0 {
			notFoundFields := make([]string, 0, length)
			for field := range entry.Data {
				if foundFieldsMap[field] == false {
					notFoundFields = append(notFoundFields, field)
				}
			}

			sort.Strings(notFoundFields)

			for i, field := range notFoundFields {
				f.writeField(b, entry, field, i)
			}
		}
		b.WriteString("]")
	}
}

func (f *Formatter) writeField(b *bytes.Buffer, entry *logrus.Entry, field string, i int) {
	if f.HideKeys {
		fmt.Fprintf(b, "%v", entry.Data[field])
	} else {
		fmt.Fprintf(b, "%s:%v", field, entry.Data[field])
	}

	if i != len(entry.Data) && len(entry.Data) != 1 {
		b.WriteString(" | ")
	}
}

func (f *Formatter) writeCaller(b *bytes.Buffer, entry *logrus.Entry) {
	if entry.HasCaller() {
		if f.CustomCallerFormatter != nil {
			fmt.Fprintf(b, f.CustomCallerFormatter(entry.Caller))
		} else {
			fmt.Fprintf(
				b,
				" (%s:%d %s)",
				entry.Caller.File,
				entry.Caller.Line,
				entry.Caller.Function,
			)
		}
	}
}
