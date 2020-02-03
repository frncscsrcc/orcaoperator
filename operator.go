package main

import (
	"flag"
	"orcaoperator/pkg/operator"
)

func main() {
	config := operator.GetDefaultConfig()

	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "kubeconfig file")
	debugLevel := flag.String("debug", "INFO", "Debug level")
	port := flag.String("port", "8012", "Webserver port")
	flag.Parse()

	config.KubeConfig = *kubeconfig
	config.DebugLevel = *debugLevel
	config.WebServerPort = *port

	operator, err := operator.New(config)
	if err != nil {
		panic(err)
	}

	operator.Init()
	operator.Run()

}
