package flow

import "fmt"

// ---

type Action interface {
	Run(TaskInfo) error
}

// --

type slackAlert struct {
	APIKey string
}

func (a slackAlert) Run(taskInfo TaskInfo) error {
	Show("------------------ This is a slack notification ------------------")
	Show("Sending alert to Slack!\n")
	Show(a.APIKey)
	Show(taskInfo.ToJSON())
	Show("------------------------------------------------------------------\n")
	return nil
}

func GetSlackAlertAction(APIKey string) (Action, error) {
	return slackAlert{APIKey: APIKey}, nil
}

// --

type sirenAlert string

func (a sirenAlert) Run(taskInfo TaskInfo) error {
	Show("------------------ This is an alarm notification ------------------")
	fmt.Printf("<<< %v %v >>>\n", a, taskInfo.Name)
	Show("-------------------------------------------------------------------\n")
	return nil
}

func GetSirenAlertAction() (Action, error) {
	return sirenAlert("UEUEUEUEUEU"), nil
}
