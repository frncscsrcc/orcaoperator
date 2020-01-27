package main

import (
	"flag"
	"fmt"
	"time"
	"orca"

	orcaV1alpha1 "orcaoperator/pkg/clients/clientset/versioned"
	orcaInformers "orcaoperator/pkg/clients/informers/externalversions"
	"k8s.io/apimachinery/pkg/labels"
//	task "orcaoperator/pkg/apis/task/v1alpha1"

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

	dataModel 			*orca.Orca
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
	orcaOperator.dataModel = orca.New()

	return orcaOperator, nil
}

func (operator *OrcaOperator) updateTaskHandler(old, new interface{}) {
	Show("CHANGED")
}

func (operator *OrcaOperator) Init() {
	taskInformer := operator.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	taskInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc : operator.updateTaskHandler,
	})
	neverStop := make(<-chan struct{})
	operator.orcaInformerFactory.Start(neverStop)
	operator.orcaInformerFactory.WaitForCacheSync(neverStop)

	tasks, err := taskInformer.Lister().Tasks("default").List(labels.Everything())
	if err != nil{
		Show(err)
	}
	dataModel := operator.dataModel
	for _, task := range(tasks) {
		name := task.ObjectMeta.Name
		dataModel.RegisterTask(name)
		t, _ := dataModel.GetTask(name)
		for _, ignitorName := range(task.Spec.StartWhen.Ignitors ) {
			dataModel.RegisterIgnitor(ignitorName)
			t.AddStartOnIgnition(ignitorName)
		}
		for _, taskName := range(task.Spec.StartWhen.Tasks.OnSuccess){
			t.AddStartOnSuccess(taskName)
		}
		for _, taskName := range(task.Spec.StartWhen.Tasks.OnFailure){
			t.AddStartOnFailure(taskName)
		}		
		Show(dataModel)
	}

}

func (operator *OrcaOperator) Run() {
	Show("RUNNING!")
}

func Show(i interface{}) {
	fmt.Printf("%+v\n", i)
}
