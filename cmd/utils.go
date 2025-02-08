package cmd

import "github.com/ginolatorilla/devops/pkg/must"

func _must[T any](v T, err error) T {
	return must.Must(v, err)
}
