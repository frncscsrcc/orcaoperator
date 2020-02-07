package operator

import (
	"time"
)

type queueItem struct {
	operation string
	item      string
}

func (qi queueItem) getOperation() string {
	return qi.operation
}

func (o *Operator) ququeIgnitorExecution(duration time.Duration, ignitorName string) {
	qi := queueItem{
		operation: "EXECUTE_IGNITOR",
		item:      ignitorName,
	}
	o.workqueue.AddAfter(qi, duration)
}

func (o *Operator) ququeTaskExecution(taskName string) {
	qi := queueItem{
		operation: "EXECUTE_TASK",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) ququeIgnitorDeletion(ignitorName string) {
	qi := queueItem{
		operation: "DELETE_IGNITOR",
		item:      ignitorName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) ququePodDeletion(podName string) {
	qi := queueItem{
		operation: "DELETE_POD",
		item:      podName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) ququeTaskStatePending(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_PENDING",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) ququeTaskStateRunning(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_RUNNING",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) ququeTaskMarkSuccess(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_SUCCESS_TIME",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) ququeTaskMarkFailure(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_FAILURE_TIME",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}
