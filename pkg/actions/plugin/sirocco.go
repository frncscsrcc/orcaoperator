package plugin

import (
	"bytes"
	"net/http"
)

type SiroccoDemoAlert struct {
	Message string
}

func (a SiroccoDemoAlert) Run(taskInfo map[string]string) error {
	url := "http://echobot.sirocco.cloud/addMessage"

	message := a.Message
	if message == "" {
		message = "GENERIC ALERT"
	}

	var jsonStr = []byte(`{"id":"` + taskInfo["Name"] + `", "message":"` + message + `"}`)
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