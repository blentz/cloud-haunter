package types

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

type TamrLicense struct {
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type TamrHealth struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

type TamrResponseBody struct {
	License    TamrLicense `json:"license"`
	TamrHealth `json:",omitempty"`
}

func ParseTamrLicenseSchema(filterName string, jsonBody []byte) TamrResponseBody {
	// function to validate the JSON schema, not the values.

	log.Debugf("[%s] JSON Body: %+v", filterName, jsonBody)
	var i TamrResponseBody
	json.Unmarshal(jsonBody, &i)
	return i
}
