package kubectlplugin

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type RunnerOpts func(*Runner)

func WithTablePrinter() RunnerOpts {
	return func(r *Runner) {
		r.ConfigFlags = NewConfigFlags()
	}
}

func WithResourcePrinters() RunnerOpts {
	return func(r *Runner) {
		r.ConfigFlags = NewConfigFlagsWithResourcePrinters()
	}
}

func WithKubeApi(kubeApi kubernetes.Interface) RunnerOpts {
	return func(r *Runner) {
		r.KubeApi = kubeApi
	}
}

func WithDefaultKubeApi() RunnerOpts {
	return func(r *Runner) {
		config, err := r.ConfigFlags.ToRESTConfig()
		if err != nil {
			panic(err)
		}
		r.KubeApi = kubernetes.NewForConfigOrDie(config)
	}
}

func WithDiscoveryApiV2(discoveryApi discovery.DiscoveryInterface, dynamicApi dynamic.Interface) RunnerOpts {
	return func(r *Runner) {
		r.DiscoveryApi = discoveryApi
		r.DynamicApi = dynamicApi
	}
}

func WithDefaultDiscoveryApi() RunnerOpts {
	return func(r *Runner) {
		config, err := r.ConfigFlags.ToRESTConfig()
		const DisableRateLimiter = -1
		config.QPS = DisableRateLimiter
		if err != nil {
			panic(err)
		}
		r.DiscoveryApi = discovery.NewDiscoveryClientForConfigOrDie(config)
		r.DynamicApi = dynamic.NewForConfigOrDie(config)
	}
}
