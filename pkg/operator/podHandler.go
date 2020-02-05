package operator

import (
	core "k8s.io/api/core/v1"
)

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

	// Stop to observer this pod
	o.podToObserve[taskName] = false

	// Handle success
	if terminated.ExitCode == 0 {
		o.log.Info.Printf("Task %s (%s) terminated with SUCCESS\n", taskName, podName)
		runTasksOnSuccess := o.flow.TriggerSuccess(taskName)
		for _, task := range runTasksOnSuccess {
			name := task.GetName()
			o.log.Info.Printf("Triggering task %s\n", name)
			o.ququeTaskExecution(name)
		}
	}

	// Handle success
	if terminated.ExitCode == 1 {
		o.log.Error.Printf("Task %s (%s) terminated with FAILURE\n", taskName, podName)
		runTasksOnFailure := o.flow.TriggerFailure(taskName)
		for _, task := range runTasksOnFailure {
			name := task.GetName()
			o.log.Info.Printf("Triggering task %s\n", name)
			o.ququeTaskExecution(name)
		}
	}

}
