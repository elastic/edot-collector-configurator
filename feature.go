package main

import (
	"fmt"
	"maps"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type FeatureParams struct {
	RootName           string
	SourceFile         []byte
	ConfigurationNames []string
}

type configuration struct {
	Content map[string]any `validate:"required"`
}

type feature struct {
	Configuration map[string]configuration `validate:"required"`
}

func BuildFeature(params FeatureParams) (map[string]any, error) {
	feature, err := parseFeatureFile(params.SourceFile)
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
		params.RootName: body,
	}, nil
}

func parseFeatureFile(data []byte) (*feature, error) {
	validate := validator.New()
	result := feature{}
	dec := yaml.NewDecoder(
		strings.NewReader(string(data)),
		yaml.Validator(validate),
		yaml.Strict(),
	)
	err := dec.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
