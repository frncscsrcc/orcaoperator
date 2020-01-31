package operator

import (
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
)

func (o *Operator) addedTaskHandler(new interface{}) {
	if !o.initialized {
		return
	}

	o.log.Trace.Println("Task added event")

	task, ok := new.(*v1alpha1.Task)
	if !ok {
		o.log.Error.Println("object is not a task")
		return
	}

	// TODO: change in RegisterTaskWithOptions
	if t, err := o.flow.RegisterTask(task.ObjectMeta.Name); err != nil {
		o.log.Error.Println("can not register task " + task.ObjectMeta.Name)
		o.log.Error.Println(err)
		return
	} else {
		o.log.Info.Println("Registered task " + task.ObjectMeta.Name)
		t.SetGeneration(task.ObjectMeta.Generation)
	}
}

func (o *Operator) updatedTaskHandler(old, new interface{}) {
	if !o.initialized {
		return
	}

	o.log.Trace.Println("Task updated event")

	task, ok := new.(*v1alpha1.Task)
	if !ok {
		o.log.Error.Println("object is not a task")
		return
	}

	taskName := task.ObjectMeta.Name

	t, err := o.flow.GetTask(taskName)
	if err != nil {
		o.log.Error.Println(err)
		return
	}

	if t.IsUpdated(task.ObjectMeta.Generation) == false {
		o.log.Trace.Println("no changes for " + taskName)
		return
	}

	o.log.Trace.Println("Updating " + taskName)

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

	if ok := o.flow.RemoveTask(name); ok {
		o.log.Info.Println("Removed task " + name)
	}
}
