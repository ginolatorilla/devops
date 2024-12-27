package cmd

import "github.com/ginolatorilla/devops/pkg/must"

var _check = must.Check

func _must[T any](v T, err error) T {
	return must.Must(v, err)
}
