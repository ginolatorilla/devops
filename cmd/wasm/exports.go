package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	gojs "syscall/js"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/ginolatorilla/devops/pkg/js"
)

func exportFuncsToJS() {
	for fname, f := range map[string]func([]gojs.Value) any{
		"renderGoTemplate": renderGoTemplate,
	} {
		jsFunc := func(this gojs.Value, args []gojs.Value) any {
			return f(args)
		}
		gojs.Global().Set(fname, gojs.FuncOf(jsFunc))
	}
	fmt.Println("Functions exported to JS")
}

func renderGoTemplate(args []gojs.Value) any {
	return js.NewPromise(func() (string, error) {
		if err := js.CheckArgs(args, gojs.TypeString, gojs.TypeString); err != nil {
			return "", err
		}
		templateExpr := args[0].String()
		tpl, err := template.New("this").Funcs(sprig.FuncMap()).Parse(templateExpr)
		if err != nil {
			return "", fmt.Errorf("first arg must be a valid Go template expression")
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(args[1].String()), &data); err != nil {
			return "", fmt.Errorf("second arg must be a JSON-formatted string")
		}
		output := bytes.Buffer{}
		if err := tpl.Execute(&output, data); err != nil {
			return "", fmt.Errorf("failed to render template: %v", err)
		}
		return output.String(), nil
	})
}
