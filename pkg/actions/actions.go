package actions

import (
	"errors"
	"orcaoperator/pkg/actions/plugin"
	"strings"
)

type Action interface {
	Run(map[string]string) error
}

// GetActionFromName returns a valid Action interface based on the action name
func GetActionFromName(action string) (Action, error) {
	if strings.ToUpper(action) == "SIROCCO-DEMO-ALERT-SUCCESS" {
		return plugin.SiroccoDemoAlert{Message: "SUCCESS"}, nil
	}
	if strings.ToUpper(action) == "SIROCCO-DEMO-ALERT-FAILURE" {
		return plugin.SiroccoDemoAlert{Message: "FAILURE"}, nil
	}
	return nil, errors.New("action " + action + " not recognized")
}
