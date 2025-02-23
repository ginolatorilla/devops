// Copyright © 2025 Gino Latorilla
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newCorsCheck() *cobra.Command {
	var (
		targetUrl string
		originUrl string
	)
	cmd := &cobra.Command{
		Use:   "cors-check --target-url URL [flags]",
		Short: "Test CORS headers",
		Long: `cors-check sends a preflight request to the target URL and checks if the response has the required headers.

See also https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return corsCheck(targetUrl, originUrl, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&targetUrl, "target-url" /* name */, "" /* default */, "The target URL to test (required)" /* usage */)
	cmd.MarkFlagRequired("target-url")
	cmd.Flags().StringVar(&originUrl, "origin-url" /* name */, "http://example.com" /* default */, "Simulate the CORS request from this URL" /* usage */)
	return cmd
}

func corsCheck(targetUrl, originUrl string, stdout io.Writer) error {
	preflight := sendPreflightRequest(targetUrl, originUrl)
	if preflight.StatusCode < 200 || preflight.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", preflight.StatusCode)
	}
	allowedMethods, ok := preflight.Header["Access-Control-Allow-Methods"]
	if !ok {
		return fmt.Errorf("Access-Control-Allow-Methods header not found")
	}
	fmt.Fprintf(stdout, "✅ %s supports CORS\n", targetUrl)
	fmt.Fprintf(stdout, "Allowed methods: %s\n", strings.Join(allowedMethods, "; "))
	fmt.Fprintf(stdout, "Allowed headers: %s\n", strings.Join(preflight.Header["Access-Control-Allow-Headers"], "; "))
	fmt.Fprintf(stdout, "Allowed origins: %s\n", strings.Join(preflight.Header["Access-Control-Allow-Origin"], "; "))
	return nil
}

func sendPreflightRequest(targetUrl, originUrl string) *http.Response {
	req := _must(http.NewRequest(http.MethodOptions, targetUrl, nil))
	req.Header.Set("Origin", originUrl)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Requested-With")
	return _must(http.DefaultClient.Do(req))
}
