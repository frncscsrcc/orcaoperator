package operator

import (
	"time"
)

type queueItem struct {
	operation string
	item      string
	initiator string
	message string
}

func (qi queueItem) getOperation() string {
	return qi.operation
}

func (o *Operator) queueIgnitorExecution(duration time.Duration, ignitorName string) {
	qi := queueItem{
		operation: "EXECUTE_IGNITOR",
		item:      ignitorName,
	}
	o.workqueue.AddAfter(qi, duration)
}

func (o *Operator) queueTaskExecution(taskName string, initiator string, message string) {
	qi := queueItem{
		operation: "EXECUTE_TASK",
		item:      taskName,
		initiator: initiator,
		message: message,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) queueIgnitorDeletion(ignitorName string) {
	qi := queueItem{
		operation: "DELETE_IGNITOR",
		item:      ignitorName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) queuePodDeletion(podName string, delay int) {
	qi := queueItem{
		operation: "DELETE_POD",
		item:      podName,
	}
	o.workqueue.AddAfter(qi, time.Duration(delay) *time.Second)
}

func (o *Operator) queueTaskStatePending(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_PENDING",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) queueTaskStateRunning(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_RUNNING",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) queueTaskMarkSuccess(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_SUCCESS_TIME",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}

func (o *Operator) queueTaskMarkFailure(taskName string) {
	qi := queueItem{
		operation: "SET_TASK_FAILURE_TIME",
		item:      taskName,
	}
	o.workqueue.AddAfter(qi, time.Duration(0))
}
