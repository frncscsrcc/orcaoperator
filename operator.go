package main

import (
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	orca "orcaoperator/pkg/clients/clientset/versioned"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	coreClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	coreAPI := coreClientSet.CoreV1()
	coreAPI.Pods("default").Get("example", metav1.GetOptions{})

	orcaClientSet, err := orca.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	orcaAPI := orcaClientSet.SiroccoV1alpha1()
	task, _ := orcaAPI.Tasks("default").Get("task1", metav1.GetOptions{})

	Show(task)

}

func Show(i interface{}) {
	fmt.Printf("%+v\n", i)
}
