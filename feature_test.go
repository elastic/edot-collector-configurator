package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var simpleConfiguration = `
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

var mergeableConfiguration = `
configuration:
  http:
    content:
      protocol:
        http:
          endpoint: http_endpoint
  grpc:
    content:
      protocol:
        grpc:
          endpoint: grpc_endpoint
`

var unmergeableConfiguration = `
configuration:
  first:
    content:
      protocol:
        http:
          endpoint: first_http_endpoint
  second:
    content:
      protocol:
        http:
          endpoint: second_http_endpoint
`

func TestBuildFeature(t *testing.T) {
	for _, tc := range []struct {
		testName             string
		input                string
		featureType          string
		configurations       []string
		expectedResult       map[string]any
		expectedErrorMessage string
		shouldFail           bool
	}{
		{
			testName:       "select configuration",
			input:          simpleConfiguration,
			featureType:    "elasticsearch",
			configurations: []string{"someconfig"},
			expectedResult: map[string]any{
				"elasticsearch": map[string]any{
					"endpoint": "someconfig_endpoint",
					"api_key":  "someconfig_api_key",
				},
			},
		},
		{
			testName:       "using default config when none provided",
			input:          simpleConfiguration,
			featureType:    "elasticsearch",
			configurations: []string{},
			expectedResult: map[string]any{
				"elasticsearch": map[string]any{
					"endpoint": "default_endpoint",
					"api_key":  "default_api_key",
				},
			},
		},
		{
			testName:       "merging configurations",
			input:          mergeableConfiguration,
			featureType:    "otlp",
			configurations: []string{"http", "grpc"},
			expectedResult: map[string]any{
				"otlp": map[string]any{
					"protocol": map[string]any{
						"http": map[string]any{
							"endpoint": "http_endpoint",
						},
						"grpc": map[string]any{
							"endpoint": "grpc_endpoint",
						},
					},
				},
			},
		},
		{
			testName:             "fail merging configurations",
			input:                unmergeableConfiguration,
			featureType:          "otlp",
			configurations:       []string{"first", "second"},
			shouldFail:           true,
			expectedErrorMessage: "key overlap for 'endpoint'",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			result, err := BuildFeature(Params{
				Type:               tc.featureType,
				SourceFileReader:   strings.NewReader(tc.input),
				ConfigurationNames: tc.configurations,
			})

			if tc.shouldFail {
				assert.EqualError(t, err, tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
