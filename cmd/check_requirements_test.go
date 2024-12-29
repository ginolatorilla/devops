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
		mockAws      exec.MockExec
		mockCurl     exec.MockExec
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
	mockExecutor.
		On("Do", context.Background(), "aws", "--version").
		Return(&mockAws)
	mockAws.
		On("Output").
		Return([]byte("aws-cli/2.22.3 Python/3.12.6 Darwin/24.1.0 exe/x86_64"), nil)
	mockExecutor.
		On("Do", context.Background(), "curl", "--version").
		Return(&mockCurl)
	mockCurl.
		On("Output").
		Return([]byte(
			`curl 8.7.1 (x86_64-apple-darwin24.0) libcurl/8.7.1 (SecureTransport) LibreSSL/3.3.6 zlib/1.2.12 nghttp2/1.62.0
Release-Date: 2024-03-27
Protocols: dict file ftp ftps gopher gophers http https imap imaps ipfs ipns ldap ldaps mqtt pop3 pop3s rtsp smb smbs smtp smtps telnet tftp
Features: alt-svc AsynchDNS GSS-API HSTS HTTP2 HTTPS-proxy IPv6 Kerberos Largefile libz MultiSSL NTLM SPNEGO SSL threadsafe UnixSockets`),
			nil,
		)

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
		mockKubectl  exec.MockExec
		mockJq       exec.MockExec
		mockOpenssl  exec.MockExec
		mockAws      exec.MockExec
		mockCurl     exec.MockExec
	)
	mockExecutor.
		On("Do", context.Background(), "bash", "--version").
		Return(&mockBash)
	mockBash.
		On("Output").
		Return(
			[]byte(`GNU bash, version 0.0.0(1)-release (aarch64-unknown-linux-gnu)
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
					"gitVersion": "v0.0.0"
				}
			}`),
			nil,
		)
	mockExecutor.
		On("Do", context.Background(), "jq", "--version").
		Return(&mockJq)
	mockJq.
		On("Output").
		Return([]byte("jq-0.0"), nil)
	mockExecutor.
		On("Do", context.Background(), "openssl", "version").
		Return(&mockOpenssl)
	mockOpenssl.
		On("Output").
		Return([]byte("OpenSSL 0.0.0 22 Oct 2024 (Library: OpenSSL 0.0.0 22 Oct 2024)"), nil)
	mockExecutor.
		On("Do", context.Background(), "aws", "--version").
		Return(&mockAws)
	mockAws.
		On("Output").
		Return([]byte("aws-cli/0.0.0 Python/3.12.6 Darwin/24.1.0 exe/x86_64"), nil)
	mockExecutor.
		On("Do", context.Background(), "curl", "--version").
		Return(&mockCurl)
	mockCurl.
		On("Output").
		Return([]byte(
			`curl 0.0.0 (x86_64-apple-darwin24.0) libcurl/8.7.1 (SecureTransport) LibreSSL/3.3.6 zlib/1.2.12 nghttp2/1.62.0
Release-Date: 2024-03-27
Protocols: dict file ftp ftps gopher gophers http https imap imaps ipfs ipns ldap ldaps mqtt pop3 pop3s rtsp smb smbs smtp smtps telnet tftp
Features: alt-svc AsynchDNS GSS-API HSTS HTTP2 HTTPS-proxy IPv6 Kerberos Largefile libz MultiSSL NTLM SPNEGO SSL threadsafe UnixSockets`),
			nil,
		)

	cmd := newCheckRequirementsCmd(mockExecutor.Executor())

	assert.Panics(t, func() { cmd.Execute() })
}

func TestCheckRequirements_List(t *testing.T) {
	t.Parallel()

	var mockExecutor exec.MockExecutor
	cmd := newCheckRequirementsCmd(mockExecutor.Executor())
	cmd.SetArgs([]string{"--list"})

	err := cmd.Execute()

	assert.NoError(t, err)
}
