package kubectlplugin

type RunnerOpts func(*Runner)

func WithResourcePrinters() RunnerOpts {
	return func(r *Runner) {
		r.configFlags = NewConfigFlagsWithResourcePrinters()
	}
}

func WithDiscoveryApi(daf DynamicApiFactory) RunnerOpts {
	return func(r *Runner) {
		r.dynamicApiFactory = daf
	}
}
