package main

import (
	"fmt"
	"io"
	"reflect"

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
	var configs = params.ConfigurationNames
	if len(configs) == 0 {
		configs = []string{"default"}
	}
	for _, key := range configs {
		configuration, ok := feature.Configuration[key]
		if !ok {
			return nil, fmt.Errorf("couldn't find configuration named '%v'", key)
		}
		err := mergeMaps(body, configuration.Content)
		if err != nil {
			return nil, err
		}
	}

	return map[string]any{
		params.Type: body,
	}, nil
}

func mergeMaps(dst map[string]any, src map[string]any) error {
	var err error = nil
	for k, v := range src {
		dstVal, found := dst[k]
		if found {
			if reflect.TypeOf(v).Kind() == reflect.Map {
				err = mergeMaps(dstVal.(map[string]any), v.(map[string]any))
			} else {
				err = fmt.Errorf("key overlap for '%v'", k)
			}
		} else {
			dst[k] = v
		}
		if err != nil {
			break
		}
	}
	return err
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
