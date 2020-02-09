package operator

import (
	"errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"math/rand"
	"orcaoperator/pkg/apis/sirocco.cloud/v1alpha1"
	"k8s.io/client-go/util/retry"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func (o *Operator) registerTasks() error {
	// Initialize the flow (based on task and ignitors already present in the cluster)
	taskInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	tasks, err := taskInformer.Lister().Tasks(o.namespace).List(labels.Everything())
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

	// Set the state to pending (TODO we should check is not runnng)
	if(task.Status.State == ""){
		o.queueTaskStatePending(task.ObjectMeta.Name)
	}

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
		if err := t.AddActionOnSuccess(actionName); err == nil {
			o.log.Trace.Println("Registered success action " + actionName + " for task " + task.ObjectMeta.Name)
		}
	}

	// Register the action to be executed in case of success
	for _, actionName := range task.Spec.FailureActions {
		if err := t.AddActionOnFailure(actionName); err == nil {
			o.log.Trace.Println("Registered failure action " + actionName + " for task " + task.ObjectMeta.Name)
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

func (o *Operator) executeTask(taskName string, initiator string, message string, done func()) error {
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

	o.queueTaskStateRunning(task.ObjectMeta.Name)

	o.log.Info.Println("Executing task " + taskName + " (background)")
	go done()

	pod := o.getPodObject(task, initiator, message)

	o.podToObserve[task.ObjectMeta.Name] = true

	pod, err = o.coreClientSet.CoreV1().Pods(o.namespace).Create(pod)
	if err != nil {
		o.log.Error.Println("Can not create a pod for " + taskName + ". Skip")
		o.log.Error.Println(err)
		done()
		return err
	}

	return nil
}

func (o *Operator) changeTaskState(taskName string, newState string, done func()) error {
	// We do not need to wait. Release the working queue
	done()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Task before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		task, err := o.getTaskByName(taskName)
		if err != nil {
			return err
		}

		task.Status.State = newState;

		_, updateErr := o.orcaClientSet.SiroccoV1alpha1().Tasks(o.namespace).Update(task)
		return updateErr
	});

	if retryErr != nil {
		return retryErr;
	}

	o.log.Info.Println("Changed state of deployment " + taskName + " in " + newState)

	return nil;
}

func (o *Operator) changeCompletedTimeState(taskName string, success bool, done func()) error {
	// We do not need to wait. Release the working queue
	done()
	
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Task before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		task, err := o.getTaskByName(taskName)
		if err != nil {
			return err
		}

		nowString := time.Now().Format("2006-01-02 15:04:05");
		if success {
			task.Status.LastSuccess = nowString
			task.Status.FailuresCount = 0
		} else {
			task.Status.LastFailure = nowString
			task.Status.FailuresCount = task.Status.FailuresCount + 1
		}

		_, updateErr := o.orcaClientSet.SiroccoV1alpha1().Tasks(o.namespace).Update(task)
		return updateErr
	});

	if retryErr != nil {
		return retryErr;
	}

	o.log.Trace.Println("Changed last complete time for " + taskName )

	return nil;
}

func (o *Operator) getTaskByName(taskName string) (*v1alpha1.Task, error) {
	tasksInformer := o.orcaInformerFactory.Sirocco().V1alpha1().Tasks()
	task, err := tasksInformer.Lister().Tasks(o.namespace).Get(taskName)
	if err != nil {
		return nil, errors.New("task " + taskName + " is not regitered in the cluster")
	}
	return task, nil
}

func (o *Operator) getPodObject(task *v1alpha1.Task, initiator string, message string) *core.Pod {
	name := task.ObjectMeta.Name
	template := task.Spec.Template

	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + getRandomString(8),
			Namespace: o.namespace,
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
					Env: []core.EnvVar{
						core.EnvVar{
							Name: "ORCA_INITIATOR",
							Value: initiator,
						},
						core.EnvVar{
							Name: "ORCA_DATA",
							Value: message,
						},
					},
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
