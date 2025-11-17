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

var configurationWithVars = `
vars:
  first: global_first
  second: false
  third: global_third
  fourth: global_fourth
configuration:
  default:
    content:
      first_placeholder: $vars.first
      second_placeholder: $vars.second
      third_placeholder: $vars.third
      some_list:
        - The fourth is $vars.fourth
      some_map_list:
        - some_key: Some value
          some_other_key: Some other value $vars.third
    vars:
      second: true
      third: config_third 
`

var configurationWithMissingVars = `
vars:
  first: global_first
configuration:
  default:
    content:
      first_placeholder: $vars.first
      second_placeholder: $vars.second
`

var configurationWithRefs = `
vars:
  first: global_first
  second: global_second
refs:
  base:
    first_placeholder: $vars.first
    second_placeholder: $vars.second
    some_key: some value
    some_other_ref: $refs.config_details
configuration:
  default:
    content: $refs.base
    refs:
      config_details:
        config_key: some config value
    vars:
      second: config_second
`

var appendingToConfiguration = `
vars:
  first: global_first
  second: global_second
refs:
  base:
    first_placeholder: $vars.first
    second_placeholder: $vars.second
    some_map:
      first_key: first value
    some_list:
      - Monday
      - Tuesday
      - Wednesday
configuration:
  default:
    content: $refs.base
    append:
      - path: "$"
        content:
          assed_to_root: root value
      - path: "$.some_map"
        content:
          added_to_child_map: child map value with $vars.second
      - path: "$.some_list"
        content:
          - Thursday
          - Friday
          - $vars.first
    vars:
      second: config_second
`

func TestBuildFeature(t *testing.T) {
	for _, tc := range []struct {
		testName             string
		input                string
		featureType          string
		configurations       []string
		vars                 map[string]any
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
		{
			testName:       "variables overriding",
			input:          configurationWithVars,
			featureType:    "dummy",
			configurations: []string{"default"},
			vars: map[string]any{
				"third":  "external_third",
				"fourth": "external_fourth",
			},
			expectedResult: map[string]any{
				"dummy": map[string]any{
					"first_placeholder":  "global_first",
					"second_placeholder": true,
					"third_placeholder":  "external_third",
					"some_list":          []any{"The fourth is external_fourth"},
					"some_map_list": []any{
						map[string]any{
							"some_key":       "Some value",
							"some_other_key": "Some other value external_third",
						},
					},
				},
			},
		},
		{
			testName:             "missing variable",
			input:                configurationWithMissingVars,
			featureType:          "dummy",
			configurations:       []string{"default"},
			expectedErrorMessage: "'$vars.second' is not defined",
			shouldFail:           true,
		},
		{
			testName:       "config with refs",
			input:          configurationWithRefs,
			featureType:    "dummy",
			configurations: []string{"default"},
			expectedResult: map[string]any{
				"dummy": map[string]any{
					"first_placeholder":  "global_first",
					"second_placeholder": "config_second",
					"some_key":           "some value",
					"some_other_ref": map[string]any{
						"config_key": "some config value",
					},
				},
			},
		},
		{
			testName:       "appending to config",
			input:          appendingToConfiguration,
			featureType:    "dummy",
			configurations: []string{"default"},
			expectedResult: map[string]any{
				"dummy": map[string]any{
					"first_placeholder":  "golbal_first",
					"second_placeholder": "config_second",
					"some_map": map[string]any{
						"first_key":          "first value",
						"added_to_child_map": "child map value with config_second",
					},
					"some_list": []any{
						"Monday",
						"Tuesday",
						"Wednesday",
						"Thrusday",
						"Friday",
						"global_first",
					},
					"assed_to_root": "root value",
				},
			},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			result, err := BuildFeature(Params{
				Type:               tc.featureType,
				SourceFileReader:   strings.NewReader(tc.input),
				ConfigurationNames: tc.configurations,
				Vars:               tc.vars,
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

func TestYamlPathParsing(t *testing.T) {
	for _, tc := range []struct {
		testName             string
		input                string
		expectedOutput       []string
		expectedErrorMessage string
		shouldFail           bool
	}{
		{
			testName:       "simple path",
			input:          "$.one.path",
			expectedOutput: []string{"one", "path"},
		},
		{
			testName:       "root path",
			input:          "$",
			expectedOutput: []string{},
		},
		{
			testName:       "path with dot",
			input:          "$.some.path.'with.dot'.other",
			expectedOutput: []string{"some", "path", "with.dot", "other"},
		},
		{
			testName:             "invalid path",
			input:                "$.",
			expectedErrorMessage: "invalid yaml path: $.",
			shouldFail:           true,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			items, err := parseYamlPath(tc.input)
			if tc.shouldFail {
				assert.EqualError(t, err, tc.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, items)
			}
		})
	}
}
