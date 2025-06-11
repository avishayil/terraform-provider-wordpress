// Copyright (c) Avishay Bar
// SPDX-License-Identifier: MIT

package provider

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockCommander mocks command execution output for testing.
type mockCommander struct {
	output []byte
	err    error
}

func (m mockCommander) CombinedOutput(name string, args ...string) ([]byte, error) {
	return m.output, m.err
}

// getTestConfig returns a WPConfig with a dynamic container name if provided.
func getTestConfig() *WPConfig {
	container := os.Getenv("WP_CONTAINER_NAME")
	if container == "" {
		container = "terraform-provider-wordpress-wordpress-1"
	}
	return &WPConfig{
		SSHTarget:  "docker:" + container + "-1",
		RemotePath: "/var/www/html",
		AllowRoot:  true,
	}
}

func TestBuildWPArgs_AllOptions(t *testing.T) {
	cfg := getTestConfig()
	args := buildWPArgs(cfg, "plugin", "install", "akismet")

	assert.Contains(t, args, "--ssh="+cfg.SSHTarget)
	assert.Contains(t, args, "--allow-root")
	assert.Contains(t, args, "--path=/var/www/html")
	assert.Equal(t, []string{
		"--ssh=" + cfg.SSHTarget,
		"--allow-root",
		"--path=/var/www/html",
		"plugin", "install", "akismet",
	}, args)
}

func TestBuildWPArgs_SomeOptions(t *testing.T) {
	cfg := &WPConfig{
		SSHTarget:  "",
		AllowRoot:  false,
		RemotePath: "/foo",
	}
	args := buildWPArgs(cfg, "theme", "install")
	assert.Equal(t, []string{"--path=/foo", "theme", "install"}, args)
}

func TestBuildWPArgs_NoOptions(t *testing.T) {
	cfg := &WPConfig{}
	args := buildWPArgs(cfg, "theme", "status")
	assert.Equal(t, []string{"theme", "status"}, args)
}

func TestRunWP_Success(t *testing.T) {
	prev := cmdExec
	defer func() { cmdExec = prev }()

	cmdExec = mockCommander{
		output: []byte("plugin installed"),
		err:    nil,
	}

	err := runWP(getTestConfig(), "plugin", "install", "akismet", "--activate")
	assert.NoError(t, err)
}

func TestRunWP_Failure(t *testing.T) {
	prev := cmdExec
	defer func() { cmdExec = prev }()

	cmdExec = mockCommander{
		output: []byte("error occurred"),
		err:    errors.New("exit code 1"),
	}

	err := runWP(getTestConfig(), "plugin", "install", "akismet")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wp")
	assert.Contains(t, err.Error(), "plugin install akismet")
	assert.Contains(t, err.Error(), "error occurred")
}

func TestRunWPWithOutput_Success(t *testing.T) {
	prev := cmdExec
	defer func() { cmdExec = prev }()

	cmdExec = mockCommander{
		output: []byte("plugin status ok"),
		err:    nil,
	}

	output, err := runWPWithOutput(getTestConfig(), "plugin", "status")
	assert.NoError(t, err)
	assert.Equal(t, "plugin status ok", output)
}

func TestRunWPWithOutput_Error(t *testing.T) {
	prev := cmdExec
	defer func() { cmdExec = prev }()

	cmdExec = mockCommander{
		output: []byte("something went wrong"),
		err:    errors.New("fail"),
	}

	output, err := runWPWithOutput(getTestConfig(), "plugin", "status")
	assert.Error(t, err)
	assert.Equal(t, "something went wrong", output)
}

func TestDefaultCommander_CombinedOutput(t *testing.T) {
	out, err := defaultCommander{}.CombinedOutput("echo", "hello")
	assert.NoError(t, err)
	assert.Contains(t, string(out), "hello")
}
