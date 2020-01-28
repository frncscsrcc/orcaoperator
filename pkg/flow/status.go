package flow

type Status string

const (
	Pending Status = "PENDING"
	Running Status = "RUNNING"
	Success Status = "SUCCESS"
	Fail    Status = "FAIL"
)
