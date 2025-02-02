package kube

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type ApiFactory func(configFlags *ConfigFlags) kubernetes.Interface

func DefaultApiFactory(configFlags *ConfigFlags) kubernetes.Interface {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	return kubernetes.NewForConfigOrDie(config)
}

type DynamicApiFactory func(configFlags *ConfigFlags) (discovery.DiscoveryInterface, dynamic.Interface)

func DefaultDynamicApiFactory(configFlags *ConfigFlags) (discovery.DiscoveryInterface, dynamic.Interface) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	return discovery.NewDiscoveryClientForConfigOrDie(config), dynamic.NewForConfigOrDie(config)
}
