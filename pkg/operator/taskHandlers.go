package operator

import (
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
	"errors"
	"k8s.io/apimachinery/pkg/labels"
)

func (o *Operator) registerTasks() error{
	// Initialize the flow (based on task and ignitors already present in the cluster)
	taskInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()	
	tasks, err := taskInformer.Lister().Tasks("default").List(labels.Everything())
	if err != nil {
		return err
	}

	flow := o.flow
	// For each existing task
	for _, task := range tasks {
		// Register the task using the name
		name := task.ObjectMeta.Name
		t, err := flow.RegisterTask(name)
		if err != nil {
			o.log.Error.Println(err)
			continue
		}
		t.SetGeneration(task.ObjectMeta.Generation)

		// Register the ignitors
		for _, ignitorName := range task.Spec.StartOnIgnition {
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

	return nil
}

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

func (o *Operator) getTaskByName(taskName string) (*v1alpha1.Task, error) {
	tasksInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	task, err := tasksInformer.Lister().Tasks("default").Get(taskName)
	if err != nil {
		return nil, errors.New("task " + taskName + " is not regitered in the cluster")
	}
	return task, nil
}

func (o  *Operator) executeTask(taskName string, done func()) error {
	// Be sure that the task at this point is still present in the cluster
	_, err := o.getTaskByName(taskName)
	if err != nil {
		return err
	}

	o.log.Info.Println("Executing task " + taskName + " (background)")
	go done()

	return nil
}
