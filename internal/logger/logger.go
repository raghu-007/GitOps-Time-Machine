// Package logger configures structured logging for GitOps-Time-Machine.
package logger

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Init configures the global logger based on configuration values.
func Init(level, format string) {
	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn", "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Set log format
	switch strings.ToLower(format) {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z",
		})
	default:
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
			ForceColors:    true,
		})
	}

	log.SetOutput(os.Stderr)
}
