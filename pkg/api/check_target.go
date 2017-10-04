package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/leapp-to/leapp-go/pkg/executor"
)

type checkTargetParams struct {
	TargetHost string `json:"target_host"`
	Status     bool   `json:"check_target_service_status"`
	TargetUser string `json:"target_user_name"`
}

func buildCheckTargetInput(p *checkTargetParams) (string, error) {
	data := map[string]interface{}{
		"target_host":                 ObjValue{p.TargetHost},
		"check_target_service_status": ObjValue{p.Status},
		"target_user_name":            ObjValue{p.TargetUser},
	}

	j, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func checkTargetHandler(request *http.Request) (interface{}, error) {
	var params checkTargetParams

	if err := json.NewDecoder(request.Body).Decode(&params); err != nil {
		return nil, err
	}

	actorInput, err := buildCheckTargetInput(&params)
	if err != nil {
		return nil, err
	}

	c := executor.New("remote-target-check-group", actorInput)
	r := c.Execute()

	log.Println(r.Stderr)

	var out interface{}
	err = json.Unmarshal([]byte(r.Stdout), &out)
	return out, err
}
