package flow

import (
	"errors"
	"orcaoperator/pkg/actions"
)

// Task reppresents a specific step in the workflow
type Task struct {
	flow       *Flow
	name       string
	generation int64

	successActions []actions.Action
	failureActions []actions.Action
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

// GetName return the task name
func (t *Task) GetName() string {
	return t.name
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

// AddActionOnSuccess register a success action, using plugins
func (t *Task) AddActionOnSuccess(actionName string) error {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()

	if action, err := actions.GetActionFromName(actionName); err == nil {
		t.successActions = append(t.successActions, action)
		return nil
	} else {
		return err
	}
}

// AddActionOnFailure register a success action, using plugins
func (t *Task) AddActionOnFailure(actionName string) error {
	t.flow.m.Lock()
	defer t.flow.m.Unlock()

	if action, err := actions.GetActionFromName(actionName); err == nil {
		t.failureActions = append(t.successActions, action)
		return nil
	} else {
		return err
	}
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
