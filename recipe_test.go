package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dummyFeature = `
vars:
  endpoint: http://localhost:8080
  api_key: default_api_key
refs:
  base:
    es_endpoint: $vars.endpoint
    es_api_key: $vars.api_key
configuration:
  default:
    content: $refs.base
  someconfig:
    content: $refs.base
    append:
      - path: "$"
        content:
          extra_key: $vars.some_var
`
var dummyRecipe = `
args:
  endpoint:
    env: ELASTICSEARCH_ENDPOINT
  api_key:
    env: ELASTICSEARCH_API_KEY
const:
  a_global_var: http://recipe.global.endpoint
features:
  my-exporter:
    source: testpath/dummy.yml
    name: custom-name
    vars:
      endpoint: $const.a_global_var
      api_key: $args.api_key
  my-other-exporter:
    source: testpath/dummy.yml
    configurations: [someconfig]
    vars:
      endpoint: $args.endpoint
      api_key: my-other-exporter-key
      some_var: other-extra-value
services:
  pipelines:
    traces:
      exporters: [ $features.my-exporter ]
    traces/something:
      exporters: [ $features.my-other-exporter ]
`

const (
	providedEndpoint = "http://external.endpoint"
	providedApiKey   = "external_api_key"
)

var expectedBuiltRecipe = `
testpath:
  dummy:
    es_api_key: my-other-exporter-key
    es_endpoint: http://external.endpoint
    extra_key: other-extra-value
  dummy/custom-name:
    es_api_key: external_api_key
    es_endpoint: http://recipe.global.endpoint
services:
  pipelines:
    traces:
      exporters: [dummy/custom-name]
    traces/something:
      exporters: [dummy]
`

func TestBuildRecipe(t *testing.T) {
	featuresTempDir, err := os.MkdirTemp("", "features")
	assert.NoError(t, err)
	os.Setenv("ELASTICSEARCH_ENDPOINT", "http://endpoint.from.env")
	os.Setenv("ELASTICSEARCH_API_KEY", providedApiKey)
	defer os.Unsetenv("ELASTICSEARCH_ENDPOINT")
	defer os.Unsetenv("ELASTICSEARCH_API_KEY")
	defer os.RemoveAll(featuresTempDir)

	testDirPath := filepath.Join(featuresTempDir, "testpath")
	err = os.Mkdir(testDirPath, 0755)
	assert.NoError(t, err)
	dummyFeatureFilePath := filepath.Join(testDirPath, "dummy.yml")
	err = os.WriteFile(dummyFeatureFilePath, []byte(dummyFeature), 0755)
	assert.NoError(t, err)

	data, err := buildRecipe(strings.NewReader(dummyRecipe), RecipeParams{
		FeaturesDirPath: featuresTempDir,
		Args: map[string]string{
			"endpoint": providedEndpoint,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expectedBuiltRecipe), string(data[:]))
}
