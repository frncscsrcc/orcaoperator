package operator

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	orcaV1alpha1 "orcaoperator/pkg/clients/clientset/versioned"
	orcaInformers "orcaoperator/pkg/clients/informers/externalversions"
	"orcaoperator/pkg/flow"
	"time"
)

type Operator struct {
	coreClientSet       *kubernetes.Clientset
	coreInformerFactory informers.SharedInformerFactory
	orcaClientSet       *orcaV1alpha1.Clientset
	orcaInformerFactory orcaInformers.SharedInformerFactory

	flow        *flow.Flow

	initialized bool
	done        chan struct{}
	log      Log
}

func New(config *restclient.Config) (*Operator, error) {
	o := &Operator{}

	// Keep a reference of the kubernetes core (pods, ...) client set
	coreClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return o, err
	}

	// Initialize a shared informer factory for the core objects (eg: pods)
	coreInformerFactory := informers.NewSharedInformerFactory(coreClientSet, time.Second*30)

	// Keep a reference of the orca specific CRD (tasks and ignitors) client set
	orcaClientSet, err := orcaV1alpha1.NewForConfig(config)
	if err != nil {
		return o, err
	}

	// Initialize a shared informer factory for the orca objects (tasks and ignitors)
	orcaInformerFactory := orcaInformers.NewSharedInformerFactory(orcaClientSet, time.Second*30)

	o.coreClientSet = coreClientSet
	o.coreInformerFactory = coreInformerFactory
	o.orcaClientSet = orcaClientSet
	o.orcaInformerFactory = orcaInformerFactory
	o.flow = flow.New()
	o.done = make(chan struct{})
	o.log = NewLog()

	return o, nil
}

func (o *Operator) Init() {
	// Initialize the task informers
	taskInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	taskInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: o.addedTaskHandler,
		UpdateFunc: o.updatedTaskHandler,
		DeleteFunc: o.deletedTaskHandler,
	})

	// Initialize the ignitor informers
	ignitorsInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Ignitors()
	ignitorsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: o.addedIgnitorHandler,
		UpdateFunc: o.updatedIgnitorHandler,
		DeleteFunc: o.deletedIgnitorHandler,		
	})

	// Activate the informer and wait for the cache
	o.orcaInformerFactory.Start(o.done)
	o.orcaInformerFactory.WaitForCacheSync(o.done)

	// Initialize the initial flow (based on task and ignitors already present in the cluster)
	tasks, err := taskInformer.Lister().Tasks("default").List(labels.Everything())
	if err != nil {
		Show(err)
		return
	}

	flow := o.flow
	// For each existing task
	for _, task := range tasks {
		Show(task.TypeMeta)
		// Register the task using the name
		name := task.ObjectMeta.Name
		flow.RegisterTask(name)
		t, _ := flow.GetTask(name)
		
		// Register the ignitors
		for _, ignitorName := range task.Spec.StartOnIgnition {
			flow.RegisterIgnitor(ignitorName)
			t.AddStartOnIgnition(ignitorName)
		}

		// Register StartOnSuccess tasks
		for _, taskName := range task.Spec.StartOnSuccess {
			t.AddStartOnSuccess(taskName)
		}

		// Register StartOnFailure tasks
		for _, taskName := range task.Spec.StartOnFailure {
			t.AddStartOnFailure(taskName)
		}
	}

	// Initialize the initial flow (based on task and ignitors already present in the cluster)
	ignitors, err := ignitorsInformer.Lister().Ignitors("default").List(labels.Everything())
	if err != nil {
		Show(err)
		return
	}

	// For each existing ignitor
	for _, ignitor := range ignitors {
		Show(ignitor)
		// Register the ignitor using the name
		name := ignitor.ObjectMeta.Name
		flow.RegisterIgnitor(name)
	}

	// Mark the fact the object is initialized
	o.initialized = true

	o.log.Info.Println("Orca is initialized")
}


func (o *Operator) Run() {
	o.log.Info.Println("Orca is running")
	// Wait done signal
	<- o.done
}

func Show(i interface{}) {
	fmt.Printf("%+v\n", i)
}
