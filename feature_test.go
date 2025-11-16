package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var simpleCase = `
configuration:
  default:
    content:
      endpoint: default_endpoint
      api_key: default_api_key
  someconfig:
    content:
      endpoint: someconfig_endpoint
      api_key: someconfig_api_key
`

func TestBuildFeature(t *testing.T) {
	for _, tc := range []struct {
		testname       string
		input          string
		featureType    string
		configurations []string
		expected       map[string]any
	}{
		{
			testname:       "select configuration",
			input:          simpleCase,
			featureType:    "elasticsearch",
			configurations: []string{"someconfig"},
			expected: map[string]any{
				"elasticsearch": map[string]any{
					"endpoint": "someconfig_endpoint",
					"api_key":  "someconfig_api_key",
				},
			},
		},
		{
			testname:       "using default config when none provided",
			input:          simpleCase,
			featureType:    "elasticsearch",
			configurations: []string{},
			expected: map[string]any{
				"elasticsearch": map[string]any{
					"endpoint": "default_endpoint",
					"api_key":  "default_api_key",
				},
			},
		},
	} {
		t.Run(tc.testname, func(t *testing.T) {
			result, err := BuildFeature(Params{
				Type:               tc.featureType,
				SourceFileReader:   strings.NewReader(tc.input),
				ConfigurationNames: tc.configurations,
			})

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
