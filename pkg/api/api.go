package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/leapp-to/leapp-go/pkg/executor"
)

// Result contains the information generated by and endpoint handler.
type Result struct {
	Errors []Error     `json:"errors,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// Error contains details about error returned by endpoint handler.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// genericResponseHandler wraps the result of the endpoint handler into a reponse that should be sent to the client.
func genericResponseHandler(fn func(*http.Request) (*executor.Result, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(writer)
		var result Result

		r, err := fn(request)
		if err != nil {
			e := fmt.Errorf("error on endpoint handler execution: %v", err)
			result.Errors = append(result.Errors, Error{1, e.Error()})
			encoder.Encode(result)
			return
		}

		// If requested, log actor's stderr
		v := request.Context().Value("Verbose").(bool)
		if v {
			log.Printf("Actor stderr: %s\n", r.Stderr)
		}

		// Runner returned an exit code different from 0
		if r.ExitCode != 0 {
			msg := fmt.Sprintf("Actor execution failed with: %d", r.ExitCode)
			result.Errors = append(result.Errors, Error{2, msg})
			encoder.Encode(result)
			return
		}

		// Actor returned an empty string (i.e. something went wrong with actor execution)
		if r.Stdout == "" {
			result.Errors = append(result.Errors, Error{3, "Actor dint't return data"})
			encoder.Encode(result)
			return
		}

		// Decode result from actor and send it back to client
		var stdout interface{}
		if err := json.Unmarshal([]byte(r.Stdout), &stdout); err != nil {
			e := fmt.Errorf("could not decode actor output: %v", err)
			result.Errors = append(result.Errors, Error{1, e.Error()})
			encoder.Encode(result)
			return
		}
		result.Data = stdout

		encoder.Encode(result)
	}
}

// EndpointEntry represents an endpoint exposed by the daemon.
type EndpointEntry struct {
	Method      string
	Endpoint    string
	HandlerFunc http.HandlerFunc
}

// GetEndpoints should return a slice of all endpoints that the daemon exposes.
func GetEndpoints() []EndpointEntry {
	return []EndpointEntry{
		{
			Method:      "POST",
			Endpoint:    "/migrate-machine",
			HandlerFunc: genericResponseHandler(migrateMachineHandler),
		},
		{
			Method:      "POST",
			Endpoint:    "/port-inspect",
			HandlerFunc: genericResponseHandler(portInspectHandler),
		},
		{
			Method:      "POST",
			Endpoint:    "/check-target",
			HandlerFunc: genericResponseHandler(checkTargetHandler),
		},
		{
			Method:      "POST",
			Endpoint:    "/port-map",
			HandlerFunc: genericResponseHandler(portMapHandler),
		},
		{
			Method:      "POST",
			Endpoint:    "/destroy-container",
			HandlerFunc: genericResponseHandler(destroyContainerHandler),
		},
	}
}
