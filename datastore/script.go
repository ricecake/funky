package datastore

import (
	"encoding/json"
	"fmt"
)

func (rec *StorageRecord) ReadScript() (string, error) {
	if rec.Type != "script" {
		return "", fmt.Errorf("Expected record type 'script', got: %s", rec.Type)
	}
	var script string
	decErr := json.Unmarshal([]byte(rec.Value), &script)
	if decErr != nil {
		return "", decErr
	}
	return script, nil
}
