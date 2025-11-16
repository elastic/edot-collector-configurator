package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var singleConfiguration = `
configuration:
  someconfig:
    content:
      endpoint: http://some.endpoint
      api_key: some_api_key
`

func TestBuildFeature(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:  "single configuration",
			input: singleConfiguration,
			expected: map[string]any{
				"elasticsearch": map[string]any{
					"endpoint": "http://some.endpoint",
					"api_key":  "some_api_key",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := BuildFeature(Params{
				Type:               "elasticsearch",
				SourceFileReader:   strings.NewReader(tc.input),
				ConfigurationNames: []string{"someconfig"},
			})

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
