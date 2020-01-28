package flow

import "errors"

// Ignitor reppresents a singnal that trigger a new workflow
type Ignitor struct {
	o    *Flow
	name string
}

// RegisterIgnitor registers a new Ignitor
func (o *Flow) RegisterIgnitor(name string) (*Ignitor, error) {
	o.m.Lock()
	defer o.m.Unlock()

	if _, exists := o.ignitors[name]; exists {
		return nil, errors.New("ignitor " + name + " exist")
	}

	i := &Ignitor{
		o:    o,
		name: name,
	}
	o.ingnitorsToTasks[name] = make(map[string]bool)
	o.ignitors[name] = i
	return i, nil
}

// RemoveIgnitor removes an Ignitor, if it was already registered. Returns true
// if the ignitor was already registered, false if it was not.
func (o *Flow) RemoveIgnitor(name string) bool {
	o.m.Lock()
	defer o.m.Unlock()

	if _, exists := o.ignitors[name]; exists {
		delete(o.ignitors, name)
		return true
	}
	return false
}
