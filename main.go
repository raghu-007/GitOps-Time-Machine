// Package main is the entry point for GitOps-Time-Machine.
package main

import (
	"os"

	"github.com/raghu-007/GitOps-Time-Machine/cmd"
	log "github.com/sirupsen/logrus"
)

// Version and BuildTime are set at build time via ldflags.
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	cmd.SetVersionInfo(Version, BuildTime)

	if err := cmd.Execute(); err != nil {
		log.WithError(err).Fatal("execution failed")
		os.Exit(1)
	}
}
