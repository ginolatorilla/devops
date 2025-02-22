package kubectlplugin

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type RunnerOpts func(*Runner)

func WithResourcePrinters() RunnerOpts {
	return func(r *Runner) {
		r.configFlags = NewConfigFlagsWithResourcePrinters()
	}
}

func WithKubeApi(kubeApi kubernetes.Interface) RunnerOpts {
	return func(r *Runner) {
		r.kubeApi = kubeApi
	}
}

func WithDefaultKubeApi() RunnerOpts {
	return func(r *Runner) {
		config, err := r.configFlags.ToRESTConfig()
		if err != nil {
			panic(err)
		}
		r.kubeApi = kubernetes.NewForConfigOrDie(config)
	}
}

func WithDiscoveryApiV2(discoveryApi discovery.DiscoveryInterface, dynamicApi dynamic.Interface) RunnerOpts {
	return func(r *Runner) {
		r.discoveryApi = discoveryApi
		r.dynamicApi = dynamicApi
	}
}

func WithDiscoveryApi(daf DynamicApiFactory) RunnerOpts {
	return func(r *Runner) {
		r.dynamicApiFactory = daf
	}
}

func WithDefaultDiscoveryApi() RunnerOpts {
	return func(r *Runner) {
		config, err := r.configFlags.ToRESTConfig()
		const DisableRateLimiter = -1
		config.QPS = DisableRateLimiter
		if err != nil {
			panic(err)
		}
		r.discoveryApi = discovery.NewDiscoveryClientForConfigOrDie(config)
		r.dynamicApi = dynamic.NewForConfigOrDie(config)
	}
}
