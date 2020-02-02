package flow

import "errors"

// TaskOptions allows to register a new task passing options (such as actions)
type TaskOptions struct {
	SuccessActions []Action
	FailureActions []Action
}

// Task reppresents a specific step in the workflow
type Task struct {
	flow       *Flow
	name       string
	generation int64

	successActions []Action
	failureActions []Action
}

// RegisterTask registers a new task in the workflow, setting only the name
func (f *Flow) RegisterTask(name string) (*Task, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if _, exists := f.tasks[name]; exists {
		return nil, errors.New("task " + name + " exists")
	}

	t := &Task{
		flow: f,
		name: name,
	}

	f.tasks[name] = t
	return t, nil
}

// RegisterTaskWithOptions registers a new task in the workflow, setting the
// name and other options.
func (f *Flow) RegisterTaskWithOptions(name string, options TaskOptions) (*Task, error) {
	t, err := f.RegisterTask(name)
	if err != nil {
		return nil, err
	}
	t.successActions = options.SuccessActions
	t.failureActions = options.FailureActions
	return t, nil
}

// RemoveTask removes a task from the workflow. Returns true if the task was present
func (f *Flow) RemoveTask(name string) bool {
	f.m.Lock()
	defer f.m.Unlock()

	f.removeTaskFromOnIgnition(name)
	f.removeTaskFromOnSuccess(name)
	f.removeTaskFromOnFailure(name)

	if _, exists := f.tasks[name]; exists {
		delete(f.tasks, name)
		return true
	}
	return false
}

// GetTask returns a pointer to a task, searched by name.
func (f *Flow) GetTask(name string) (*Task, error) {
	t, exists := f.tasks[name]
	if exists {
		return t, nil
	}
	return nil, errors.New("task " + name + " not found")
}

// AddStartOnIgnition sets an ignitor for the task
func (t *Task) AddStartOnIgnition(ignitorName string) {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()

	if _, exists := t.flow.ingnitorsToTasks[ignitorName]; !exists {
		t.flow.ingnitorsToTasks[ignitorName] = make(map[string]bool)
	}

	t.flow.ingnitorsToTasks[ignitorName][t.name] = true
}

// AddStartOnSuccess set a listener for the current task, that will start if
// task name terminates with success
func (t *Task) AddStartOnSuccess(taskName string) {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()

	if _, exists := t.flow.successToTasks[taskName]; !exists {
		t.flow.successToTasks[taskName] = make(map[string]bool)
	}

	t.flow.successToTasks[taskName][t.name] = true
}

// AddStartOnSuccess set a listener for the current task, that will start if
// task name terminates with a failure
func (t *Task) AddStartOnFailure(taskName string) {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()
	if _, exists := t.flow.failureToTasks[taskName]; !exists {
		t.flow.failureToTasks[taskName] = make(map[string]bool)
	}

	t.flow.failureToTasks[taskName][t.name] = true
}

// SetGeneration updates the task generation (version)
func (t *Task) SetGeneration(generation int64) {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()
	t.generation = generation
}

// IsUpdated checks if a task definition is updated
func (t *Task) IsUpdated(generation int64) bool {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()
	return generation > t.generation
}

func (f *Flow) removeTaskFromOnIgnition(name string) bool {
	for _, taskMap := range f.ingnitorsToTasks {
		if _, exists := taskMap[name]; exists {
			delete(taskMap, name)
			return true
		}
	}
	return false
}

func (f *Flow) removeTaskFromOnSuccess(name string) bool {
	for _, taskMap := range f.successToTasks {
		if _, exists := taskMap[name]; exists {
			delete(taskMap, name)
			return true
		}
	}
	return false
}

func (f *Flow) removeTaskFromOnFailure(name string) bool {
	for _, taskMap := range f.failureToTasks {
		if _, exists := taskMap[name]; exists {
			delete(taskMap, name)
			return true
		}
	}
	return false
}
