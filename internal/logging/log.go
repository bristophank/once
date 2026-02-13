package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func SetupFile() (func(), error) {
	dir, err := stateDir()
	if err != nil {
		return nil, fmt.Errorf("determining log directory: %w", err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	path := filepath.Join(dir, "once.log")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(file, nil)))

	return func() { file.Close() }, nil
}

func SetupStderr() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
}

// Helpers

func stateDir() (string, error) {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, "once"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "state", "once"), nil
}
