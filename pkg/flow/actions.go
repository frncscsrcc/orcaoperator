package flow

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
)

// ---

type Action interface {
	Run(TaskInfo) error
}

// --

type siroccoDemoAlert struct {
	message string
}

func (a siroccoDemoAlert) Run(taskInfo TaskInfo) error {
	url := "http://echobot.sirocco.cloud/addMessage"

	message := a.message
	if message == "" {
		message = "GENERIC ALERT"
	}

	var jsonStr = []byte(`{"id":"` + taskInfo.Name + `", "message":"` + message + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func getActionFromName(action string) (Action, error) {
	if strings.ToUpper(action) == "SIROCCO-DEMO-ALERT-SUCCESS" {
		return siroccoDemoAlert{message: "SUCCESS"}, nil
	}
	if strings.ToUpper(action) == "SIROCCO-DEMO-ALERT-FAILURE" {
		return siroccoDemoAlert{message: "FAILURE"}, nil
	}
	return nil, errors.New("action " + action + " not recognized")
}
