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
type refs map[string]map[string]any

type append struct {
	Path    string
	Content any
}

type configuration struct {
	Content any `validate:"required"`
	Vars    vars
	Refs    refs
	Append  append
}

type feature struct {
	Configuration map[string]configuration `validate:"required"`
	Vars          vars
	Refs          refs
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
		configRefs := collectRefs(feature, configuration)
		configContent, err := resolveConfigContent(configuration.Content, configRefs)
		if err != nil {
			return nil, err
		}
		err = mergeMaps(body, configContent)
		if err != nil {
			return nil, err
		}

		configVars := collectVars(feature, configuration, params)

		err = replaceVarsInMap(body, configVars)
		if err != nil {
			return nil, err
		}
	}

	return map[string]any{
		params.Type: body,
	}, nil
}

var refsPattern = regexp.MustCompile(`^\$refs\.[^\s]+$`)

func resolveConfigContent(content any, configRefs refs) (map[string]any, error) {
	if isMap(content) {
		err := resolveMapRefs(content.(map[string]any), configRefs)
		if err != nil {
			return nil, err
		}
		return content.(map[string]any), nil
	} else if isString(content) {
		mapRef, err := resolveStringRef(content.(string), configRefs)
		if err != nil {
			return nil, err
		}
		err = resolveMapRefs(mapRef, configRefs)
		if err != nil {
			return nil, err
		}
		return mapRef, nil
	}
	return nil, fmt.Errorf("invalid content type, must be a map or a ref to a map. It's: %v", content)
}

func resolveMapRefs(content map[string]any, configRefs refs) error {
	for k, v := range content {
		if isString(v) && refsPattern.MatchString(v.(string)) {
			mapRef, err := resolveStringRef(v.(string), configRefs)
			if err != nil {
				return err
			}
			err = resolveMapRefs(mapRef, configRefs)
			if err != nil {
				return err
			}
			content[k] = mapRef
		}
	}
	return nil
}

func resolveStringRef(content string, configRefs refs) (map[string]any, error) {
	refId := refsPattern.FindString(content)
	if refId == "" {
		return nil, fmt.Errorf("'%v' is not a valid ref", content)
	}
	ref, ok := configRefs[refId]
	if !ok {
		return nil, fmt.Errorf("'%s' is not defined", refId)
	}
	return ref, nil
}

func collectRefs(feature *feature, configuration configuration) refs {
	collected := maps.Clone(feature.Refs)
	maps.Copy(collected, configuration.Refs)

	refPrefixedMap := make(refs, len(collected))
	for k, v := range collected {
		refPrefixedMap["$refs."+k] = v
	}

	return refPrefixedMap
}

func collectVars(feature *feature, configuration configuration, params Params) vars {
	collected := maps.Clone(feature.Vars)
	maps.Copy(collected, configuration.Vars)
	maps.Copy(collected, params.Vars)

	varPrefixedMap := make(vars, len(collected))
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
	for k, v := range src {
		dstVal, found := dst[k]
		if found {
			if isMap(v) {
				err := mergeMaps(dstVal.(map[string]any), v.(map[string]any))
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("key overlap for '%v'", k)
			}
		} else {
			dst[k] = v
		}
	}
	return nil
}

func isMap(value any) bool {
	return reflect.TypeOf(value).Kind() == reflect.Map
}

func isList(value any) bool {
	kind := reflect.TypeOf(value).Kind()
	return kind == reflect.Slice || kind == reflect.Array
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
