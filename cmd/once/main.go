package main

import (
	"os"

	"github.com/basecamp/once/internal/command"
	"github.com/basecamp/once/internal/logging"
)

func main() {
	logging.SetupStderr()

	if err := command.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
