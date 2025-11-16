package main

import (
	"fmt"
	"io"
	"maps"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type Params struct {
	Type               string
	SourceFileReader   io.Reader
	ConfigurationNames []string
}

func NewParams(sourceFilePath string, configurationNames []string) Params {
	panic("implement")
}

type configuration struct {
	Content map[string]any `validate:"required"`
}

type feature struct {
	Configuration map[string]configuration `validate:"required"`
}

func BuildFeature(params Params) (map[string]any, error) {
	feature, err := parseFeatureFile(params.SourceFileReader)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any)
	for i := range params.ConfigurationNames {
		key := params.ConfigurationNames[i]
		configuration, ok := feature.Configuration[key]
		if !ok {
			return nil, fmt.Errorf("couldn't find configuration named '%v'", key)
		}
		maps.Copy(body, configuration.Content)
	}

	return map[string]any{
		params.Type: body,
	}, nil
}

func parseFeatureFile(data io.Reader) (*feature, error) {
	validate := validator.New()
	result := feature{}
	dec := yaml.NewDecoder(
		data,
		yaml.Validator(validate),
		yaml.Strict(),
	)
	err := dec.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
