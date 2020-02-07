package flow

import (
	"sync"
)

// Options ...
type Options struct{}

// Flow handles the internal representation of the the process graph
type Flow struct {
	m sync.Mutex

	ingnitorsToTasks map[string](map[string]bool)
	successToTasks   map[string](map[string]bool)
	failureToTasks   map[string](map[string]bool)

	tasks    map[string]*Task
	ignitors map[string]*Ignitor
}

// New initializes an Flow object.
func New() *Flow {
	return &Flow{
		ingnitorsToTasks: make(map[string](map[string]bool)),
		successToTasks:   make(map[string](map[string]bool)),
		failureToTasks:   make(map[string](map[string]bool)),
		tasks:            make(map[string]*Task),
		ignitors:         make(map[string]*Ignitor),
	}
}

// NewWithOptions initializes an Flow object and sets internal options
func NewWithOptions(Options) *Flow {
	f := New()
	return f
}

func (f *Flow) TriggerIgnition(ignitionName string) []*Task {
	f.m.Lock()
	defer f.m.Unlock()
	tasks := make([]*Task, 0)

	var taskNamesMap map[string]bool
	var exists bool

	if taskNamesMap, exists = f.ingnitorsToTasks[ignitionName]; !exists {
		return tasks
	}

	for taskName := range taskNamesMap {
		task, err := f.GetTask(taskName)
		if err == nil {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (f *Flow) TriggerSuccess(taskName string) []*Task {
	f.m.Lock()
	defer f.m.Unlock()

	// If task exists, trigger the onSuccessActions
	if t, err := f.GetTask(taskName); err == nil {
		taskInfo := map[string]string{"Name": taskName}
		for _, action := range t.successActions {
			action.Run(taskInfo)
		}
	}

	tasks := make([]*Task, 0)

	var taskNamesMap map[string]bool
	var exists bool

	if taskNamesMap, exists = f.successToTasks[taskName]; !exists {
		return tasks
	}

	for taskName := range taskNamesMap {
		task, err := f.GetTask(taskName)
		if err == nil {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (f *Flow) TriggerFailure(taskName string) []*Task {
	f.m.Lock()
	defer f.m.Unlock()

	// If task exists, trigger the onSuccessActions
	if t, err := f.GetTask(taskName); err == nil {
		taskInfo := map[string]string{"Name": taskName}
		for _, action := range t.failureActions {
			action.Run(taskInfo)
		}
	}

	tasks := make([]*Task, 0)

	var taskNamesMap map[string]bool
	var exists bool

	if taskNamesMap, exists = f.failureToTasks[taskName]; !exists {
		return tasks
	}

	for taskName := range taskNamesMap {
		task, err := f.GetTask(taskName)
		if err == nil {
			tasks = append(tasks, task)
		}
	}

	return tasks
}
