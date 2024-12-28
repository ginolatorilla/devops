package cmd

import (
	"context"
	"testing"

	"github.com/ginolatorilla/devops/pkg/exec"
	"github.com/stretchr/testify/assert"
)

func TestCheckRequirements(t *testing.T) {
	t.Parallel()
	var (
		mockExecutor exec.MockExecutor
		mockBash     exec.MockExec
		mockKubectl  exec.MockExec
		mockJq       exec.MockExec
		mockOpenssl  exec.MockExec
	)
	mockExecutor.
		On("Do", context.Background(), "bash", "--version").
		Return(&mockBash)
	mockBash.
		On("Output").
		Return(
			[]byte(`GNU bash, version 5.2.21(1)-release (aarch64-unknown-linux-gnu)
Copyright (C) 2022 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>

This is free software; you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.`),
			nil,
		)
	mockExecutor.
		On("Do", context.Background(), "kubectl", "version", "--client", "--output", "json").
		Return(&mockKubectl)
	mockKubectl.
		On("Output").
		Return(
			[]byte(`{
				"clientVersion": {
					"gitVersion": "v1.29.0"
				}
			}`),
			nil,
		)
	mockExecutor.
		On("Do", context.Background(), "jq", "--version").
		Return(&mockJq)
	mockJq.
		On("Output").
		Return([]byte("jq-1.7"), nil)
	mockExecutor.
		On("Do", context.Background(), "openssl", "version").
		Return(&mockOpenssl)
	mockOpenssl.
		On("Output").
		Return([]byte("OpenSSL 3.4.0 22 Oct 2024 (Library: OpenSSL 3.4.0 22 Oct 2024)"), nil)

	cmd := newCheckRequirementsCmd(mockExecutor.Executor())

	err := cmd.Execute()

	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
	mockBash.AssertExpectations(t)
	mockKubectl.AssertExpectations(t)
	mockJq.AssertExpectations(t)
	mockOpenssl.AssertExpectations(t)
}

func TestCheckRequirements_PanicIfNotMet(t *testing.T) {
	t.Parallel()
	var (
		mockExecutor exec.MockExecutor
		mockBash     exec.MockExec
	)
	mockExecutor.
		On("Do", context.Background(), "bash", "--version").
		Return(&mockBash)
	mockBash.
		On("Output").
		Return(
			[]byte(`GNU bash, version 1.0.0(1)-release (aarch64-unknown-linux-gnu)
Copyright (C) 2022 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>

This is free software; you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.`),
			nil,
		)
	cmd := newCheckRequirementsCmd(mockExecutor.Executor())

	assert.Panics(t, func() { cmd.Execute() })
	mockExecutor.AssertExpectations(t)
	mockBash.AssertExpectations(t)
}

func TestCheckRequirements_List(t *testing.T) {
	t.Parallel()

	var mockExecutor exec.MockExecutor
	cmd := newCheckRequirementsCmd(mockExecutor.Executor())
	cmd.SetArgs([]string{"--list"})

	err := cmd.Execute()

	assert.NoError(t, err)
}
