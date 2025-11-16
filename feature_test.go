package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var inputBasic = `
configuration:
  someconfig:
    content:
      endpoint: http://some.endpoint
      api_key: some_api_key
`

func TestBuildFeature(t *testing.T) {
	for _, tc := range []struct {
		input    string
		expected map[string]any
	}{
		{
			input: inputBasic,
			expected: map[string]any{
				"elasticsearch": map[string]any{
					"endpoint": "http://some.endpoint",
					"api_key":  "some_api_key",
				},
			},
		},
	} {
		result, err := BuildFeature(Params{
			Type:               "elasticsearch",
			SourceFileReader:   strings.NewReader(tc.input),
			ConfigurationNames: []string{"someconfig"},
		})

		assert.NoError(t, err)
		assert.Equal(t, tc.expected, result)
	}
}
