package gocbt

import (
	"strings"
)

func geopoint(path, name string) map[string]interface{} {
	objs := strings.Split(path, ".")

	properties := make(map[string]interface{})
	current := properties
	for i, obj := range objs {
		if i+1 < len(objs) {
			props := make(map[string]interface{})
			current[obj] = map[string]interface{}{
				"dynamic":    false,
				"enabled":    true,
				"properties": props,
			}
			current = props
		} else {
			current[obj] = map[string]interface{}{
				"dynamic": false,
				"enabled": true,
				"fields": []interface{}{
					map[string]interface{}{
						"index":                true,
						"name":                 name,
						"store":                true,
						"type":                 "geopoint",
						"include_in_all":       true,
						"include_term_vectors": true,
					},
				},
			}
		}
	}

	return properties
}
