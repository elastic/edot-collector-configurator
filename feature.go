package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

type FetureParams struct {
	Type               string
	Name               string
	ConfigurationNames []string
	Vars               map[string]any
}

type varsType map[string]any
type refsType map[string]map[string]any

type appendType struct {
	Path    string `validate:"required"`
	Content any    `validate:"required"`
}

type configurationType struct {
	Content any `validate:"required"`
	Vars    varsType
	Refs    refsType
	Append  []appendType
}

type featureType struct {
	Configuration map[string]configurationType `validate:"required"`
	Vars          varsType
	Refs          refsType
}

const varsPatternStr = `\$vars\.[^\s]+`

var (
	refsPattern          = regexp.MustCompile(`^\$refs\.[^\s]+$`)
	varsPattern          = regexp.MustCompile(varsPatternStr)
	fullStringVarPattern = regexp.MustCompile(fmt.Sprintf("^%s$", varsPatternStr))
	yamlPathPattern      = regexp.MustCompile(`^\$((?:\.[^\s.]+)+)?$`)
	dotSeparatedPattern  = regexp.MustCompile(`'[^\s]+'|[^.\s]+`)
)

func BuildFeature(sourceFilePath string, params FetureParams) (map[string]any, error) {
	f, err := os.Open(sourceFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return buildFeature(f, params)
}

func buildFeature(source io.Reader, params FetureParams) (map[string]any, error) {
	var err error
	feature, err := parseFeatureFile(source)
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
		err = appendItems(body, configuration.Append)
		if err != nil {
			return nil, err
		}
		configVars, err := collectVars(feature, configuration, params)
		if err != nil {
			return nil, err
		}

		err = replaceVarsInMap(body, configVars)
		if err != nil {
			return nil, err
		}
	}

	featureName := params.Type
	if len(params.Name) > 0 {
		featureName = fmt.Sprintf("%s/%s", featureName, params.Name)
	}
	return map[string]any{
		featureName: body,
	}, nil
}

func appendItems(body map[string]any, appendType []appendType) error {
	var err error
	for _, item := range appendType {
		err = appendItem(body, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func appendItem(body map[string]any, item appendType) error {
	var err error
	path, err := parseYamlPath(item.Path)
	if err != nil {
		return err
	}
	if isMap(item.Content) {
		err = appendMapItems(body, path, item.Content.(map[string]any))
		if err != nil {
			return err
		}
	} else if isList(item.Content) {
		err = appendListItems(body, path, item.Content.([]any))
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid append content type, must be a map or list - it's: %v", getKind(item.Content))
	}
	return nil
}

func appendMapItems(body map[string]any, path []string, content map[string]any) error {
	var targetMap map[string]any = body
	var ok bool
	for _, pathItem := range path {
		targetMap, ok = targetMap[pathItem].(map[string]any)
		if !ok {
			return fmt.Errorf("could not find item '%s' via yaml path: %v", pathItem, path)
		}
	}
	for k, v := range content {
		if targetMap[k] != nil {
			return fmt.Errorf("key '%s' already exists in target map, cannot append existing keys", k)
		}
		targetMap[k] = v
	}
	return nil
}

func appendListItems(body map[string]any, path []string, content []any) error {
	var targetMap map[string]any = body
	var pathToMap = path[:len(path)-1]
	var ok bool
	for _, pathItem := range pathToMap {
		targetMap, ok = targetMap[pathItem].(map[string]any)
		if !ok {
			return fmt.Errorf("could not find item '%s' via yaml path: %v", pathItem, path)
		}
	}
	listKey := path[len(path)-1]
	originalList := targetMap[listKey].([]any)
	targetMap[listKey] = append(originalList, content...)

	return nil
}

func resolveConfigContent(content any, configRefs refsType) (map[string]any, error) {
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
	return nil, fmt.Errorf("invalid content type, must be a map or a ref to a map - it's: %v", getKind(content))
}

func resolveMapRefs(content map[string]any, configRefs refsType) error {
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

func resolveStringRef(content string, configRefs refsType) (map[string]any, error) {
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

func collectRefs(feature *featureType, configuration configurationType) refsType {
	collected := maps.Clone(feature.Refs)
	maps.Copy(collected, configuration.Refs)

	refPrefixedMap := make(refsType, len(collected))
	for k, v := range collected {
		refPrefixedMap["$refs."+k] = v
	}

	return refPrefixedMap
}

func collectVars(feature *featureType, configuration configurationType, params FetureParams) (varsType, error) {
	collected := maps.Clone(feature.Vars)
	maps.Copy(collected, configuration.Vars)
	maps.Copy(collected, params.Vars)

	varPrefixedMap := make(varsType, len(collected))
	for k, v := range collected {
		varName := "$vars." + k
		if !isPrimitive(v) {
			return nil, fmt.Errorf("'%s' format is not valid, only primitives are allowed", varName)
		}
		varPrefixedMap[varName] = v
	}

	return varPrefixedMap, nil
}

func replaceVarsInMap(body map[string]any, configVars varsType) error {
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

func replaceVarsInList(list []any, configVars varsType) ([]any, error) {
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

func resolveVarsInString(value string, configVars varsType) (any, error) {
	if fullStringVarPattern.MatchString(value) {
		varValue, ok := configVars[value]
		if ok {
			return varValue, nil
		} else {
			return nil, fmt.Errorf("'%s' is not defined", value)
		}
	} else if varsPattern.MatchString(value) {
		matches := varsPattern.FindAllString(value, -1)
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
	return getKind(value) == reflect.Map
}

func isList(value any) bool {
	kind := getKind(value)
	return kind == reflect.Slice || kind == reflect.Array
}

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

func getKind(value any) reflect.Kind {
	return reflect.TypeOf(value).Kind()
}

func parseFeatureFile(data io.Reader) (*featureType, error) {
	validate := validator.New()
	result := featureType{}
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

func parseYamlPath(path string) ([]string, error) {
	match := yamlPathPattern.FindAllStringSubmatch(path, -1)
	if len(match) == 0 {
		return nil, fmt.Errorf("invalid yaml path: %s", path)
	}
	subpath := match[0][1]
	if subpath == "" {
		return []string{}, nil
	}
	var items []string
	for _, item := range dotSeparatedPattern.FindAllString(subpath, -1) {
		items = append(items, strings.Trim(item, "'"))
	}
	return items, nil
}
