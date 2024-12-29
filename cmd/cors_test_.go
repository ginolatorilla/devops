package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newCorsTestCmd() *cobra.Command {
	var (
		targetUrl string
		originUrl string
	)
	cmd := &cobra.Command{
		Use:   "cors-test --target-url URL [flags]",
		Short: "Test CORS headers",
		Long: `cors-test sends a preflight request to the target URL and checks if the response has the required headers.

See also https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request`,
		Run: func(cmd *cobra.Command, args []string) {
			corsTest(targetUrl, originUrl, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&targetUrl, "target-url" /* name */, "" /* default */, "The target URL to test (required)" /* usage */)
	cmd.MarkFlagRequired("target-url")
	cmd.Flags().StringVar(&originUrl, "origin-url" /* name */, "http://example.com" /* default */, "Simulate the CORS request from this URL" /* usage */)
	return cmd
}

func corsTest(targetUrl, originUrl string, stdout io.Writer) {
	preflight := sendPreflightRequest(targetUrl, originUrl)
	if preflight.StatusCode < 200 || preflight.StatusCode >= 300 {
		panic(fmt.Errorf("unexpected status code: %d", preflight.StatusCode))
	}
	allowedMethods, ok := preflight.Header["Access-Control-Allow-Methods"]
	if !ok {
		panic(fmt.Errorf("Access-Control-Allow-Methods header not found"))
	}
	fmt.Fprintf(stdout, "✅ %s supports CORS\n", targetUrl)
	fmt.Fprintf(stdout, "Allowed methods: %s\n", strings.Join(allowedMethods, "; "))
	fmt.Fprintf(stdout, "Allowed headers: %s\n", strings.Join(preflight.Header["Access-Control-Allow-Headers"], "; "))
	fmt.Fprintf(stdout, "Allowed origins: %s\n", strings.Join(preflight.Header["Access-Control-Allow-Origin"], "; "))
}

func sendPreflightRequest(targetUrl, originUrl string) *http.Response {
	req := _must(http.NewRequest(http.MethodOptions, targetUrl, nil))
	req.Header.Set("Origin", originUrl)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-Requested-With")
	return _must(http.DefaultClient.Do(req))
}
