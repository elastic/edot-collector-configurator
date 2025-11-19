package main

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type varsType map[string]any

func isPrimitive(value any) bool {
	kindName := getKind(value).String()
	for _, primitiveName := range []string{
		"string",
		"bool",
		"int",
		"float",
	} {
		if strings.HasPrefix(kindName, primitiveName) {
			return true
		}
	}
	return false
}

func isString(value any) bool {
	return getKind(value) == reflect.String
}

func isMap(value any) bool {
	return getKind(value) == reflect.Map
}

func isList(value any) bool {
	kind := getKind(value)
	return kind == reflect.Slice || kind == reflect.Array
}

func getKind(value any) reflect.Kind {
	return reflect.TypeOf(value).Kind()
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
			if isMap(v) {
				newMap := make(map[string]any)
				err := mergeMaps(newMap, v.(map[string]any))
				if err != nil {
					return err
				}
				dst[k] = newMap
			} else {
				dst[k] = v
			}
		}
	}
	return nil
}

func replacePlaceholdersInMap(target map[string]any, placeholderPattern regexp.Regexp, values map[string]any) error {
	for k, v := range target {
		if isMap(v) {
			err := replacePlaceholdersInMap(v.(map[string]any), placeholderPattern, values)
			if err != nil {
				return err
			}
		} else if isList(v) {
			list, err := replacePlaceholdersInList(v.([]any), placeholderPattern, values)
			if err != nil {
				return err
			}
			target[k] = list
		} else if isString(v) {
			resolvedValue, err := resolvePlaceholdersInString(v.(string), placeholderPattern, values)
			if err != nil {
				return err
			}
			target[k] = resolvedValue
		}
	}
	return nil
}

func replacePlaceholdersInList(list []any, placeholderPattern regexp.Regexp, values map[string]any) ([]any, error) {
	resolvedList := make([]any, len(list))
	for i, v := range list {
		if isMap(v) {
			err := replacePlaceholdersInMap(v.(map[string]any), placeholderPattern, values)
			if err != nil {
				return nil, err
			}
			resolvedList[i] = v
		} else if isString(v) {
			resolvedValue, err := resolvePlaceholdersInString(v.(string), placeholderPattern, values)
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

func resolvePlaceholdersInString(target string, placeholderPattern regexp.Regexp, values map[string]any) (any, error) {
	var fullTextPatterns []string
	for _, pattern := range strings.Split(placeholderPattern.String(), "|") {
		fullTextPatterns = append(fullTextPatterns, fmt.Sprintf("^%s$", pattern))
	}
	fullTextPattern := regexp.MustCompile(strings.Join(fullTextPatterns, "|"))
	if fullTextPattern.MatchString(target) {
		mapValue, ok := values[target]
		if ok {
			return mapValue, nil
		} else {
			return nil, fmt.Errorf("'%s' is not defined", target)
		}
	} else if placeholderPattern.MatchString(target) {
		matches := placeholderPattern.FindAllString(target, -1)
		if len(matches) > 0 {
			newValue := target
			for _, v := range slices.Compact(matches) {
				mapValue, ok := values[v].(string)
				if ok {
					newValue = strings.ReplaceAll(newValue, v, mapValue)
				} else {
					return nil, fmt.Errorf("'%s' is not defined", v)
				}
			}
			return newValue, nil
		}
	}
	return target, nil
}

func prependToKeysOfPrimitiveValues[V any](target map[string]V, prefix string) (map[string]V, error) {
	refPrefixedMap := make(map[string]V, len(target))
	var keyName string
	for k, v := range target {
		keyName = prefix + k
		if !isPrimitive(v) {
			return nil, fmt.Errorf("'%s' format is not valid, only primitives are allowed", keyName)
		}
		refPrefixedMap[keyName] = v
	}
	return refPrefixedMap, nil
}

func parseYamlFile(data io.Reader, result any) error {
	validate := validator.New()
	dec := yaml.NewDecoder(
		data,
		yaml.Validator(validate),
		yaml.Strict(),
	)
	err := dec.Decode(result)
	if err != nil {
		return err
	}
	return nil
}
