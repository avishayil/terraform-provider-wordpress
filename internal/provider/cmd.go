// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"os/exec"
)

// Commander is an abstraction for executing external commands.
type Commander interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

// defaultCommander uses os/exec for real command execution.
type defaultCommander struct{}

// CombinedOutput runs the command and returns combined stdout and stderr.
func (defaultCommander) CombinedOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// cmdExec can be mocked in tests, otherwise uses the real executor.
var cmdExec Commander = defaultCommander{}

// buildWPArgs assembles wp-cli arguments consistently.
func buildWPArgs(cfg *WPConfig, args ...string) []string {
	allArgs := []string{}

	if cfg.SSHTarget != "" {
		allArgs = append(allArgs, "--ssh="+cfg.SSHTarget)
	}

	if cfg.AllowRoot {
		allArgs = append(allArgs, "--allow-root")
	}

	if cfg.RemotePath != "" {
		allArgs = append(allArgs, "--path="+cfg.RemotePath)
	}

	allArgs = append(allArgs, args...)
	return allArgs
}

// runWP runs a wp-cli command and returns only an error (used for Create, Delete, Activate).
func runWP(cfg *WPConfig, args ...string) error {
	allArgs := buildWPArgs(cfg, args...)
	output, err := cmdExec.CombinedOutput("wp", allArgs...)
	if err != nil {
		return fmt.Errorf("wp %v failed: %s", allArgs, string(output))
	}
	return nil
}

// runWPWithOutput runs a wp-cli command and returns both output and error (used for status checks).
func runWPWithOutput(cfg *WPConfig, args ...string) (string, error) {
	allArgs := buildWPArgs(cfg, args...)
	output, err := cmdExec.CombinedOutput("wp", allArgs...)
	return string(output), err
}
