package types

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

type TamrVersion struct {
	Version        string        `json:"version"`
	GitDescribe    string        `json:"gitDescribe"`
	GitCommitId    string        `json:"gitCommitId"`
	GitCommitShort string        `json:"gitCommitShort"`
	GitCommitTime  TamrTimestamp `json:"gitCommitTime"`
	BuildTime      TamrTimestamp `json:"buildTime"`
}

type TamrTimestamp struct {
	Time time.Time
}

// UnmarshalJSON decodes an int64 timestamp into a time.Time object
func (p *TamrTimestamp) UnmarshalJSON(bytes []byte) error {
	// 1. Decode the bytes into a string
	var raw string
	err := json.Unmarshal(bytes, &raw)

	if err != nil {
		log.Errorf("error decoding timestamp: %s", err)
		return err
	}

	// 2 - Parse the timestamp; see: src/pkg/time/format.go
	*&p.Time, err = time.Parse("2006-01-02 15:04:05 PM MST", raw)
	if err != nil {
		log.Errorf("error decoding timestamp: %s", err)
		return err
	}

	return nil
}

func ParseTamrVersionSchema(filterName string, jsonBody []byte) TamrVersion {
	// function to validate the JSON schema, not the values.

	log.Debugf("[%s] JSON Body: %+v", filterName, jsonBody)
	var i TamrVersion
	json.Unmarshal(jsonBody, &i)
	return i
}
