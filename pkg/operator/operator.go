package operator

import (
	"fmt"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	//	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
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

	podToObserve map[string]bool

	workqueue workqueue.DelayingInterface

	flow *flow.Flow

	initialized bool
	done        chan struct{}
	log         Log
	config      Config
}

func New(appConfig Config) (*Operator, error) {
	o := &Operator{}

	config, err := clientcmd.BuildConfigFromFlags("", appConfig.KubeConfig)
	if err != nil {
		panic(err)
	}

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
	o.workqueue = workqueue.NewDelayingQueue()
	o.podToObserve = make(map[string]bool)
	o.flow = flow.New()
	o.done = make(chan struct{})
	o.log = NewLog(appConfig.DebugLevel)
	o.config = appConfig

	return o, nil
}

func (o *Operator) Init() {
	// Initialize the pod (running task) informers
	podInformer := o.coreInformerFactory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: o.updatedPodHandler,
	})

	// Initialize the task informers
	taskInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	taskInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.addedTaskHandler,
		UpdateFunc: o.updatedTaskHandler,
		DeleteFunc: o.deletedTaskHandler,
	})

	// Initialize the ignitor informers
	ignitorsInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Ignitors()
	ignitorsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.addedIgnitorHandler,
		UpdateFunc: o.updatedIgnitorHandler,
		DeleteFunc: o.deletedIgnitorHandler,
	})

	// Activate the informer and wait for the cache
	o.coreInformerFactory.Start(o.done)
	o.coreInformerFactory.WaitForCacheSync(o.done)
	o.orcaInformerFactory.Start(o.done)
	o.orcaInformerFactory.WaitForCacheSync(o.done)

	// Register the tasks already present in the cluster
	if err := o.registerTasks(); err != nil {
		o.log.Error.Println("Problem in task initialization. Skip it.")
	}

	// Register the ingitors already present in the cluster
	if err := o.registerIgnitors(); err != nil {
		o.log.Error.Println("Problem in ingitor initialization. Skip it.")
	}

	// Mark the fact the object is initialized
	o.initialized = true

	o.log.Info.Println("Orca is initialized")

	// Initialize webserver (in a separate thread)
	go o.initializeWebServer()
}

func (o *Operator) Run() {
	o.log.Info.Println("Orca is running")

	// Be sure the done channel triggers a shutdown
	// This function is execute in a separate thread
	go func(o *Operator) {
		// Wait done signal
		<-o.done
		o.workqueue.ShutDown()
	}(o)

	for true {
		generic, shutdown := o.workqueue.Get()
		if shutdown {
			break
		}

		var ok bool
		var qi queueItem

		// Cast the item to be a queueItem structure
		qi, ok = generic.(queueItem)
		if !ok {
			o.log.Info.Println("Invalid working queue item. Just ignoring it.")
			continue
		}

		o.log.Trace.Println("Received queue item " + qi.operation + " for " + qi.item)

		// Callback function to call when the queue item is processed
		itemDone := func() {
			o.workqueue.Done(qi)
		}

		switch qi.getOperation() {

		case "EXECUTE_IGNITOR":
			o.executeIgnitor(qi.item, itemDone)

		case "EXECUTE_TASK":
			{
				o.executeTask(qi.item, itemDone)
			}

		case "DELETE_IGNITOR":
			{
				o.deleteIgnitor(qi.item, itemDone)
			}
		default:

		}
	}
}

func Show(i interface{}) {
	fmt.Printf("%+v\n", i)
}
