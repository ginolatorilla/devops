package kubectlplugin

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

func As[T any](object runtime.Object) T {
	converted, ok := object.(T)
	if !ok {
		panic(fmt.Errorf("failed to convert runtime.Object to %T", converted))
	}
	return converted
}
