package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/dustin/go-humanize"
	"github.com/fatih/structs"
	"github.com/spf13/cobra"
)

var funcMap = template.FuncMap{
	"x509":         parseX509ToMap,
	"relativeTime": humanize.Time,
}

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Render a template",
		Long: `template reads a Go template from stdin and renders it to stdout.

The following functions are available in the template:
- Sprig functions (see http://masterminds.github.io/sprig/)
- x509: Parse a PEM-encoded X.509 certificate and return a map of its fields (see https://pkg.go.dev/crypto/x509#Certificate)
- relativeTime: Convert a time.Time to a human-readable relative time`,
		Run: func(cmd *cobra.Command, args []string) {
			render(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	return cmd
}

func render(stdin io.Reader, stdout io.Writer) {
	_check(_must(template.
		New("stdin").
		Funcs(sprig.FuncMap()).
		Funcs(funcMap).
		Parse(
			string(_must(io.ReadAll(stdin))),
		)).
		Execute(stdout, nil))
}

func parseX509ToMap(input string) map[string]any {
	der, _ := pem.Decode([]byte(input))
	cert := _must(x509.ParseCertificate(der.Bytes))
	return structs.Map(cert)
}
