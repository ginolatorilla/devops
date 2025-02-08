package kubectlplugin

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TabularHandler func(args HandlerArgs) (metaV1.Table, error)

func NewTabularRunner(apiFactory ApiFactory, handler TabularHandler) *Runner {
	return &Runner{
		configFlags:    NewConfigFlags(),
		apiFactory:     apiFactory,
		tabularHandler: handler,
	}
}

func NewTabularRunnerWithDiscoveryApi(apiFactory DynamicApiFactory, handler TabularHandler) *Runner {
	return &Runner{
		configFlags:       NewConfigFlags(),
		dynamicApiFactory: apiFactory,
		tabularHandler:    handler,
	}
}
