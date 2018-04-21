package datastore

import (
	"encoding/json"
	"log"
)

type RouteData map[string][]string

func LookupRoute(host string, path string) (routes []string, err error) {
	var records []StorageRecord
	GetDb().Where(`type = 'route' and key = ? and value->? is not null`, host, path).Find(&records)
	log.Printf("%+v\n", records)
	for _, record := range records {
		var data RouteData
		decodeErr := json.Unmarshal([]byte(record.Value), &data)
		if decodeErr != nil {
			return nil, decodeErr
		}
		routes = append(routes, data[path]...)
	}

	return routes, nil
}
