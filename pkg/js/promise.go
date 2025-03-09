package js

import "syscall/js"

func NewPromise[T any](fn func() (T, error)) any {
	handler := js.FuncOf(func(_ js.Value, args []js.Value) any {
		go func() {
			resolve := args[0]
			reject := args[1]
			result, err := fn()
			if err != nil {
				jsError := js.Global().Get("Error").New(err.Error())
				reject.Invoke(jsError)
				return
			}
			resolve.Invoke(result)
		}()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}
