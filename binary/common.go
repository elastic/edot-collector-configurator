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
	return isSlice(value) || isArray(value)
}

func isSlice(value any) bool {
	return getKind(value) == reflect.Slice
}

func isArray(value any) bool {
	return getKind(value) == reflect.Array
}

func getKind(value any) reflect.Kind {
	return reflect.TypeOf(value).Kind()
}

func deepCopy[T any](v T) T {
	return deepCopyAny(v).(T)
}

func deepCopyAny(value any) any {
	switch {
	case isMap(value):
		mapValue := value.(map[string]any)
		cp := make(map[string]any, len(mapValue))
		for k, v2 := range mapValue {
			cp[k] = deepCopyAny(v2)
		}
		return cp

	case isSlice(value):
		sliceValue := value.([]any)
		cp := make([]any, len(sliceValue))
		for i, v2 := range sliceValue {
			cp[i] = deepCopyAny(v2)
		}
		return cp

	default:
		return value
	}
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
			} else if isList(v) {
				dst[k] = append(dstVal.([]any)[:], v.([]any)...)
			} else {
				return fmt.Errorf("key overlap for '%v'", k)
			}
		} else {
			dst[k] = v
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
