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
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	return cmd
}

func render(stdin io.Reader, stdout io.Writer) error {
	tpl, err := template.
		New("stdin").
		Funcs(sprig.FuncMap()).
		Funcs(funcMap).
		Parse(
			string(_must(io.ReadAll(stdin))),
		)
	if err != nil {
		return err
	}
	if err := tpl.Execute(stdout, nil); err != nil {
		return err
	}
	return nil
}

func parseX509ToMap(input string) map[string]any {
	der, _ := pem.Decode([]byte(input))
	cert := _must(x509.ParseCertificate(der.Bytes))
	return structs.Map(cert)
}
