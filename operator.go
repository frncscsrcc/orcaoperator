package main

import (
	"flag"
	"orcaoperator/pkg/operator"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	operator, err := operator.New(config)
	if err != nil {
		panic(err)
	}

	operator.Init()
	operator.Run()

}

