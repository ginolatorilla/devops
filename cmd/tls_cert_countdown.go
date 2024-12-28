package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

const _iso8601 = "2006-01-02 15:04:05Z"

func newTlsCertCountdownCmd() *cobra.Command {
	var short bool
	cmd := &cobra.Command{
		Use:   "tls-cert-countdown",
		Short: "Check the expiration date of TLS certificates",
		Long: fmt.Sprintf(`tls-cert-countdown prints how much time is left until a TLS certificate expires.
This command reads the output of the openssl x509 command from stdin.

For example:
	$ openssl x509 -text -noout -dates -dateopt iso_8601 -in /path/to/cert.pem | %s tls-cert-countdown`,
			AppName,
		),
		Run: func(cmd *cobra.Command, args []string) {
			checkTlsCertCountdown(cmd.InOrStdin(), cmd.OutOrStdout(), short)
		},
	}
	cmd.Flags().BoolVar(&short, "short", false, "Print a short version of the output")
	return cmd
}

func checkTlsCertCountdown(stdin io.Reader, stdout io.Writer, short bool) {
	input := string(_must(io.ReadAll(stdin)))
	for _, line := range strings.Split(input, "\n") {
		if strings.Contains(line, "notAfter=") {
			printReport(stdout, parseOpenSslDateLine(line), short)
			return
		}
	}

	panic(fmt.Errorf("required line not found in the input: notAfter="))
}

func parseOpenSslDateLine(line string) time.Time {
	parts := strings.Split(line, "=")
	return _must(time.Parse(_iso8601, parts[1]))
}

func printReport(out io.Writer, expDate time.Time, short bool) {
	friendlyTime := humanize.Time(expDate)
	if short {
		fmt.Fprintln(out, friendlyTime)
		return
	}

	if strings.Contains(friendlyTime, "ago") {
		fmt.Fprintf(out, "The certificate expired %s\n", friendlyTime)
	} else {
		fmt.Fprintf(out, "The certificate will expire in %s\n", friendlyTime)
	}
}
