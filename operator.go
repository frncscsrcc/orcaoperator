package main

import (
	"flag"
	"fmt"
	"time"

	orcaV1alpha1 "orcaoperator/pkg/clients/clientset/versioned"
	orcaInformers "orcaoperator/pkg/clients/informers/externalversions"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	operator, err := New(config)
	if err != nil {
		panic(err)
	}

	operator.Init()
	operator.Run()

	time.Sleep(122 * time.Second)

}

type OrcaOperator struct {
	coreClientSet       *kubernetes.Clientset
	coreInformerFactory informers.SharedInformerFactory
	orcaClientSet       *orcaV1alpha1.Clientset
	orcaInformerFactory orcaInformers.SharedInformerFactory
}

func New(config *restclient.Config) (*OrcaOperator, error) {
	orcaOperator := &OrcaOperator{}

	// Keep a reference of the kubernetes core (pods, ...) client set
	coreClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return orcaOperator, err
	}

	// Initialize a shared informer factory for the core objects (eg: pods)
	coreInformerFactory := informers.NewSharedInformerFactory(coreClientSet, time.Second*30)

	// Keep a reference of the orca specific CRD (tasks and ignitors) client set
	orcaClientSet, err := orcaV1alpha1.NewForConfig(config)
	if err != nil {
		return orcaOperator, err
	}

	// Initialize a shared informer factory for the orca objects (tasks and ignitors)
	orcaInformerFactory := orcaInformers.NewSharedInformerFactory(orcaClientSet, time.Second*30)

	orcaOperator.coreClientSet = coreClientSet
	orcaOperator.coreInformerFactory = coreInformerFactory
	orcaOperator.orcaClientSet = orcaClientSet
	orcaOperator.orcaInformerFactory = orcaInformerFactory

	return orcaOperator, nil
}

func (operator *OrcaOperator) Init() {
	taskInformer := operator.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	taskInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc : func(old, new interface{}){Show("UPDATEEEEEE!")},
	})
	neverStop := make(<-chan struct{})
	operator.orcaInformerFactory.Start(neverStop)
	operator.orcaInformerFactory.WaitForCacheSync(neverStop)
}

func (operator *OrcaOperator) Run() {

}

func Show(i interface{}) {
	fmt.Printf("%+v\n", i)
}
