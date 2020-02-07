package operator

import (
	"errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"math/rand"
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func (o *Operator) registerTasks() error {
	// Initialize the flow (based on task and ignitors already present in the cluster)
	taskInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	tasks, err := taskInformer.Lister().Tasks("default").List(labels.Everything())
	if err != nil {
		return err
	}

	// For each existing task
	for _, task := range tasks {
		o.registerTask(task)
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

	o.registerTask(task)
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
	o.registerTask(task)
}

func (o *Operator) registerTask(task *v1alpha1.Task) error {
	// Remove task, in case it was present already (update)
	o.flow.RemoveTask(task.ObjectMeta.Name)

	// TODO: change in RegisterTaskWithOptions
	t, err := o.flow.RegisterTask(task.ObjectMeta.Name)
	if err != nil {
		o.log.Error.Println("can not register task " + task.ObjectMeta.Name)
		o.log.Error.Println(err)
		return err
	} else {
		o.log.Info.Println("Registered task " + task.ObjectMeta.Name)
		t.SetGeneration(task.ObjectMeta.Generation)
	}

	// Register the action to be executed in case of success
	for _, actionName := range task.Spec.SuccessActions {
		if err := t.RegisterSuccessAction(actionName); err == nil {
			o.log.Info.Println("Registered success action " + actionName + " for task " + task.ObjectMeta.Name)
		}
	}

	// Register the action to be executed in case of success
	for _, actionName := range task.Spec.FailureActions {
		if err := t.RegisterFailureAction(actionName); err == nil {
			o.log.Info.Println("Registered failure action " + actionName + " for task " + task.ObjectMeta.Name)
		}
	}

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

	return nil
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

func (o *Operator) executeTask(taskName string, done func()) error {
	// Be sure that the task at this point is still present in the cluster
	task, err := o.getTaskByName(taskName)
	if err != nil {
		o.log.Error.Println("The task " + taskName + " is not registered. Skipping")
		done()
		return err
	}

	if podRunning, exists := o.podToObserve[taskName]; exists && podRunning {
		o.log.Warning.Println("Skipping task " + taskName + " (a pod is already running in the cluster)")
		done()
		return nil
	}

	o.log.Info.Println("Executing task " + taskName + " (background)")
	go done()

	pod := o.getPodObject(task)

	o.podToObserve[task.ObjectMeta.Name] = true

	pod, err = o.coreClientSet.CoreV1().Pods("default").Create(pod)
	if err != nil {
		o.log.Error.Println("Can not create a pod for " + taskName + ". Skip")
		o.log.Error.Println(err)
		done()
		return err
	}

	return nil
}

func (o *Operator) getPodObject(task *v1alpha1.Task) *core.Pod {
	name := task.ObjectMeta.Name
	template := task.Spec.Template

	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + getRandomString(8),
			Namespace: "default",
			Labels: map[string]string{
				"app": "demo",
			},
			Annotations: map[string]string{
				"orcaTask": task.ObjectMeta.Name,
			},
		},
		Spec: core.PodSpec{
			RestartPolicy: "Never",
			Containers: []core.Container{
				{
					Name:            template.Spec.Containers[0].Name,
					Image:           template.Spec.Containers[0].Image,
					ImagePullPolicy: core.PullIfNotPresent,
					Command:         template.Spec.Containers[0].Command,
				},
			},
		},
	}
}

func getRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
