//	Copyright © 2025 Gino Latorilla
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
package main

import (
	"bytes"
	"fmt"
	gojs "syscall/js"

	"text/template"

	"github.com/ginolatorilla/devops/pkg/js"
)

func main() {
	quit := make(chan any)
	fmt.Println("Hello from WASM")
	for fname, f := range map[string]func([]gojs.Value) any{
		"renderGoTemplate": renderGoTemplate,
	} {
		jsFunc := func(this gojs.Value, args []gojs.Value) any {
			return f(args)
		}
		gojs.Global().Set(fname, gojs.FuncOf(jsFunc))
	}
	fmt.Println("Functions exported to JS")
	<-quit
}

func renderGoTemplate(args []gojs.Value) any {
	return js.NewPromise(func() (string, error) {
		if len(args) < 1 {
			return "", fmt.Errorf("expected at least 1 argument")
		}
		input := args[0].String()
		tpl, err := template.New("this").Parse(input)
		if err != nil {
			return "", err
		}
		output := bytes.Buffer{}
		if err := tpl.Execute(&output, nil); err != nil {
			return "", nil
		}
		return output.String(), nil
	})
}
