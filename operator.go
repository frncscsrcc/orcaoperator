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
	deleteSuccessPodDelay := flag.Int("deleteSuccessPodDelay", 60, "Delete POD delay on success (sec)")
	deleteFailedPodDelay := flag.Int("deleteFailedPodDelay", 300, "Delete POD delay on failure (sec)")
	keepPods := flag.Bool("keepPods", false, "Do not delete pods after termination")

	flag.Parse()

	config.KubeConfig = *kubeconfig
	config.DebugLevel = *debugLevel
	config.WebServerPort = *port
	config.DeleteSuccessPodDelay = *deleteSuccessPodDelay
	config.DeleteFailedPodDelay = *deleteFailedPodDelay
	config.KeepPods = *keepPods

	operator, err := operator.New(config)
	if err != nil {
		panic(err)
	}

	operator.Init()
	operator.Run()
}