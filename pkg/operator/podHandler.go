package operator

import (
	"errors"
	core "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (o *Operator) registerPods() error {
	// Initialize the flow (based on task and ignitors already present in the cluster)
	podInformer := o.coreInformerFactory.Core().V1().Pods()
	// TODO: filter based on orcaTask annotation
	pods, err := podInformer.Lister().Pods("default").List(labels.Everything())
	if err != nil {
		return err
	}

	// For each existing pod
	for _, pod := range pods {
		taskName := pod.ObjectMeta.Annotations["orcaTask"]

		// Consider only pod for tasks
		if taskName == "" {
			continue
		}

		// Consider only running pods
		if len(pod.Status.ContainerStatuses) == 0 {
			continue
		}
		// Check if the pod is terminated
		if pod.Status.ContainerStatuses[0].State.Terminated != nil {
			if !o.config.KeepPods{
				// Cleaning: delete terminated pod from the cluster
				o.ququePodDeletion(pod.ObjectMeta.Name, 0)
			}
			continue
		}

		o.podToObserve[taskName] = true
		o.log.Info.Println("Registered running pod " + pod.ObjectMeta.Name + " for task " + taskName)
	}

	return nil
}

func (o *Operator) updatedPodHandler(old, new interface{}) {
	if !o.initialized {
		return
	}

	podPrev, ok := old.(*core.Pod)
	if !ok {
		o.log.Error.Println("object is not a pod")
		return
	}

	pod, ok := new.(*core.Pod)
	if !ok {
		o.log.Error.Println("object is not a pod")
		return
	}

	podName := pod.ObjectMeta.Name
	taskName := pod.ObjectMeta.Annotations["orcaTask"]

	// Skip if the pos is not related to an orca task
	if taskName == "" {
		return
	}

	// Consider only pod that were generated from orca operator
	if toObserve, exists := o.podToObserve[taskName]; !exists || !toObserve {
		return
	}
	o.log.Trace.Println("Pod updated event")

	if podPrev.Status.Phase != pod.Status.Phase {
		o.log.Info.Printf("Task %s (%s) is %v\n", taskName, podName, pod.Status.Phase)
	}

	if len(pod.Status.ContainerStatuses) == 0 {
		return
	}

	// Check if the pod is terminated
	terminated := pod.Status.ContainerStatuses[0].State.Terminated
	if terminated == nil {
		return
	}

	// Retrive any termination message (e.g.: data to send to the next task)
	message := terminated.Message

	// The task is finished! Mark it as pending
	o.ququeTaskStatePending(taskName)

	// Stop to observer this pod
	o.podToObserve[taskName] = false

	// Handle success
	if terminated.ExitCode == 0 {
		o.log.Info.Printf("Task %s (%s) terminated with SUCCESS\n", taskName, podName)
		
		// Request the deletion of the pod
		if !o.config.KeepPods{
			o.ququePodDeletion(podName, o.config.DeleteSuccessPodDelay)
		}

		// Save last success time
		o.ququeTaskMarkSuccess(taskName)

		// Start all the new tasks that depend on the success of this one
		runTasksOnSuccess := o.flow.TriggerSuccess(taskName)
		for _, task := range runTasksOnSuccess {
			name := task.GetName()
			o.log.Info.Printf("Triggering task %s\n", name)
			o.ququeTaskExecution(name, taskName, message)
		}
	}

	// Handle success
	if terminated.ExitCode == 1 {
		o.log.Error.Printf("Task %s (%s) terminated with FAILURE\n", taskName, podName)

		// Request the deletion of the pod
		if !o.config.KeepPods {
			o.ququePodDeletion(podName, o.config.DeleteFailedPodDelay)
		}

		// Save last failure time
		o.ququeTaskMarkFailure(taskName)

		// Start all the new tasks that depend on the failure of this one
		runTasksOnFailure := o.flow.TriggerFailure(taskName)
		for _, task := range runTasksOnFailure {
			name := task.GetName()
			o.log.Info.Printf("Triggering task %s\n", name)
			o.ququeTaskExecution(name, taskName, message)
		}
	}
}

func (o *Operator) deletePod(podName string, done func()) error {
	// Be sure that the ignitor at this point is still present in the cluster
	_, err := o.getPodByName(podName)
	if err != nil {
		return err
	}

	if err := o.coreClientSet.CoreV1().Pods("default").Delete(podName, &metaV1.DeleteOptions{}); err != nil {
		return err
	}

	o.log.Info.Printf("Deleted pod %s\n", podName)

	done()
	return nil
}

func (o *Operator) getPodByName(podName string) (*core.Pod, error) {
	podInformer := o.coreInformerFactory.Core().V1().Pods()
	pod, err := podInformer.Lister().Pods("default").Get(podName)
	if err != nil {
		return nil, errors.New("pod " + podName + " is not regitered in the cluster")
	}
	return pod, nil
}
