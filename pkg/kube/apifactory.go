package kube

import (
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
