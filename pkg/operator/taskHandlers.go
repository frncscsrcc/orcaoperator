package operator

import (
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
)

func (o *Operator) addedTaskHandler(new interface{}) {
	if !o.initialized {
		return
	}
	o.log.Trace.Println("Task added event")
}


func (o *Operator) updatedTaskHandler(old, new interface{}) {
	if !o.initialized {
		return
	}
	o.log.Trace.Println("Task updated event")
}


func (o *Operator) deletedTaskHandler(old interface{}) {
	if !o.initialized {
		return
	}
	
	o.log.Trace.Println("Task deleted event")
	task, ok := old.(*v1alpha1.Task)

	if !ok {
		o.log.Error.Println("object is not a task")
		return
	}

	name := task.ObjectMeta.Name
	Show(name)
	if ok := o.flow.RemoveTask(name); ok {
		o.log.Info.Println("Removed task " + name)
	}
}
