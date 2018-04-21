package datastore

import (
	"encoding/json"
)

type DomainRouteData map[string][]string
type EventRouteData map[string]string

func LookupDomainRoute(host string, path string) (routes []string, owner string, scope string, err error) {
	var record StorageRecord
	GetDb().Where(`type = 'domain_route' and key = ? and value->? is not null`, host, path).Find(&record)
	if record.Id != "" {
		var data DomainRouteData
		decodeErr := json.Unmarshal([]byte(record.Value), &data)
		if decodeErr != nil {
			return routes, owner, scope, decodeErr
		}
		routes = append(routes, data[path]...)
	}

	return routes, record.Owner, record.Scope, nil
}

func LookupEventRoute(owner string, scope string, event string) (scriptRecord StorageRecord, err error) {
	var record StorageRecord
	GetDb().Where(`type = 'event_route' and owner = ? and scope <@ ? and key = ?`, owner, scope, event).Find(&record)

	if record.Id != "" {
		var data EventRouteData
		decodeErr := json.Unmarshal([]byte(record.Value), &data)
		if decodeErr != nil {
			return record, decodeErr
		}
		scriptRecord.Owner = owner
		scriptRecord.Scope = scope
		scriptRecord.Type = "script"
		scriptRecord.Key = data["script"]
	}

	return scriptRecord, nil
}
