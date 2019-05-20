package gocbt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeopoint(t *testing.T) {
	props := geopoint("one.two.three", "mygeo")

	require.Equal(t, map[string]interface{}{
		"one": map[string]interface{}{
			"dynamic": false,
			"enabled": true,
			"properties": map[string]interface{}{
				"two": map[string]interface{}{
					"dynamic": false,
					"enabled": true,
					"properties": map[string]interface{}{
						"three": map[string]interface{}{
							"dynamic": false,
							"enabled": true,
							"fields": []interface{}{
								map[string]interface{}{
									"index":                true,
									"name":                 "mygeo",
									"store":                true,
									"type":                 "geopoint",
									"include_in_all":       true,
									"include_term_vectors": true,
								},
							},
						},
					},
				},
			},
		},
	}, props)
}
