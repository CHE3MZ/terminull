package termrender

import (
	"fmt"
	"strings"
)

// LogStyle selects how flat log lines are decorated.
type LogStyle uint8

const (
	// LogStyleTree keeps the live tree on terminals and plain lines elsewhere.
	LogStyleTree LogStyle = iota
	// LogStylePlain prints undecorated flat lines.
	LogStylePlain
	// LogStyleTime prefixes each line with [15:04:05].
	LogStyleTime
	// LogStyleTimeLevel prefixes each line with [15:04:05] INFO.
	LogStyleTimeLevel
	// LogStyleLogrus prefixes each line with INFO[0000], using elapsed
	// seconds and logrus's four-character level names (ERRO, DEBU, TRAC).
	LogStyleLogrus
	// LogStyleKubernetes prefixes each line with I1026 15:04:05.123456.
	LogStyleKubernetes
)

const logStyleChoices = "tree|plain|time|time-level|logrus|k8s"

// String returns the flag value for a log style.
func (style LogStyle) String() string {
	switch style {
	case LogStyleTree:
		return "tree"
	case LogStylePlain:
		return "plain"
	case LogStyleTime:
		return "time"
	case LogStyleTimeLevel:
		return "time-level"
	case LogStyleLogrus:
		return "logrus"
	case LogStyleKubernetes:
		return "k8s"
	default:
		return "unknown"
	}
}

// ParseLogStyle parses a log style flag value.
func ParseLogStyle(value string) (LogStyle, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "tree":
		return LogStyleTree, nil
	case "plain":
		return LogStylePlain, nil
	case "time":
		return LogStyleTime, nil
	case "time-level", "clock":
		return LogStyleTimeLevel, nil
	case "logrus":
		return LogStyleLogrus, nil
	case "k8s", "kubernetes":
		return LogStyleKubernetes, nil
	default:
		return 0, fmt.Errorf("want %s", logStyleChoices)
	}
}
