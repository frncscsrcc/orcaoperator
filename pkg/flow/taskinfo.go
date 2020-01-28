package flow

import (
	"encoding/json"
	"time"
)

// TaskInfo reppresent exported fields from a tas, it can be used from Actions
type TaskInfo struct {
	Name           string
	Status         Status
	IsSuccess      bool
	StartTime      time.Time
	EndTime        time.Time
	SuccessMessage string
	ErrorMessage   string
}

func (t *Task) GetTaskInfo() TaskInfo {
	return TaskInfo{
		Name: t.name,
	}
}

func (ti TaskInfo) ToJSON() string {
	bytes, err := json.Marshal(ti)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
