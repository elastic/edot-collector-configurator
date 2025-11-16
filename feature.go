package main

import (
	"fmt"
	"io"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type Params struct {
	Type               string
	SourceFileReader   io.Reader
	ConfigurationNames []string
	Vars               map[string]any
}

func NewParams(sourceFilePath string, configurationNames []string) Params {
	panic("implement")
}

type vars map[string]any

type configuration struct {
	Content map[string]any `validate:"required"`
	Vars    vars
}

type feature struct {
	Configuration map[string]configuration `validate:"required"`
	Vars          vars
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

		configVars := collectVars(feature, configuration, params)

		replaceVarsInMap(body, configVars)
	}

	return map[string]any{
		params.Type: body,
	}, nil
}

func collectVars(feature *feature, configuration configuration, params Params) vars {
	collected := maps.Clone(feature.Vars)
	maps.Copy(collected, configuration.Vars)
	maps.Copy(collected, params.Vars)

	varPrefixedMap := make(map[string]any, len(collected))
	for k, v := range collected {
		varPrefixedMap["$vars."+k] = v
	}

	return varPrefixedMap
}

func replaceVarsInMap(body map[string]any, configVars vars) error {
	for k, v := range body {
		if isMap(v) {
			err := replaceVarsInMap(v.(map[string]any), configVars)
			if err != nil {
				return err
			}
		} else if isList(v) {
			list, err := replaceVarsInList(v.([]any), configVars)
			if err != nil {
				return err
			}
			body[k] = list
		} else if isString(v) {
			resolvedValue, err := resolveVarsInString(v.(string), configVars)
			if err != nil {
				return err
			}
			body[k] = resolvedValue
		}
	}
	return nil
}

func replaceVarsInList(list []any, configVars vars) ([]any, error) {
	resolvedList := make([]any, len(list))
	for i, v := range list {
		if isMap(v) {
			err := replaceVarsInMap(v.(map[string]any), configVars)
			if err != nil {
				return nil, err
			}
			resolvedList[i] = v
		} else if isString(v) {
			resolvedValue, err := resolveVarsInString(v.(string), configVars)
			if err != nil {
				return nil, err
			}
			resolvedList[i] = resolvedValue
		} else {
			resolvedList[i] = v
		}
	}
	return resolvedList, nil
}

func resolveVarsInString(value string, configVars vars) (any, error) {
	varsPatternStr := `\$vars\.[^\s]+`
	varPattern := regexp.MustCompile(varsPatternStr)
	fullStringVarPattern := regexp.MustCompile(fmt.Sprintf("^%s$", varsPatternStr))

	if fullStringVarPattern.MatchString(value) {
		varValue, ok := configVars[value]
		if ok {
			return varValue, nil
		} else {
			return nil, fmt.Errorf("'%s' is not defined", value)
		}
	} else if varPattern.MatchString(value) {
		matches := varPattern.FindAllString(value, -1)
		if len(matches) > 0 {
			newValue := value
			for _, v := range slices.Compact(matches) {
				varValue, ok := configVars[v].(string)
				if ok {
					newValue = strings.ReplaceAll(newValue, v, varValue)
				} else {
					return nil, fmt.Errorf("'%s' is not defined", v)
				}
			}
			return newValue, nil
		}
	}
	return value, nil
}

func mergeMaps(dst map[string]any, src map[string]any) error {
	var err error = nil
	for k, v := range src {
		dstVal, found := dst[k]
		if found {
			if isMap(v) {
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

func isMap(value any) bool {
	return reflect.TypeOf(value).Kind() == reflect.Map
}

func isList(value any) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

func isString(value any) bool {
	return reflect.TypeOf(value).Kind() == reflect.String
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
