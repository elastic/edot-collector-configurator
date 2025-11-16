package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFeature(t *testing.T) {
	yml := `
configuration:
  someconfig:
    content:
      endpoint: http://some.endpoint
      api_key: some_api_key
`
	expectedOutput := map[string]any{
		"elasticsearch": map[string]any{
			"endpoint": "http://some.endpoint",
			"api_key":  "some_api_key",
		},
	}
	result, err := BuildFeature(Params{
		Type:               "elasticsearch",
		SourceFileReader:   strings.NewReader(yml),
		ConfigurationNames: []string{"someconfig"},
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, result)
}
