package flow

import "errors"

// Ignitor reppresents a singnal that trigger a new workflow
type Ignitor struct {
	flow       *Flow
	name       string
	generation int64
}

// RegisterIgnitor registers a new Ignitor
func (f *Flow) RegisterIgnitor(name string) (*Ignitor, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if _, exists := f.ignitors[name]; exists {
		return nil, errors.New("ignitor " + name + " exist")
	}

	i := &Ignitor{
		flow: f,
		name: name,
	}
	f.ingnitorsToTasks[name] = make(map[string]bool)
	f.ignitors[name] = i
	return i, nil
}

// RemoveIgnitor removes an Ignitor, if it was already registered. Returns true
// if the ignitor was already registered, false if it was not.
func (f *Flow) RemoveIgnitor(name string) bool {
	f.m.Lock()
	defer f.m.Unlock()

	if _, exists := f.ignitors[name]; exists {
		delete(f.ignitors, name)
		return true
	}
	return false
}

// GetIgnitor returns a pointer to an ignitor, searched by name.
func (f *Flow) GetIgnitor(name string) (*Ignitor, error) {
	i, exists := f.ignitors[name]
	if exists {
		return i, nil
	}
	return nil, errors.New("ignitor " + name + " not found")
}

// SetGeneration updates the task generation (version)
func (i *Ignitor) SetGeneration(generation int64) {
	i.flow.m.Lock()
	defer i.flow.m.Unlock()
	i.generation = generation
}

// IsUpdated checks if a task definition is updated
func (i *Ignitor) IsUpdated(generation int64) bool {
	i.flow.m.Lock()
	defer i.flow.m.Unlock()
	return generation > i.generation
}
