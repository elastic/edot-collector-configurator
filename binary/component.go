package main

import (
	"fmt"
	"io"
	"maps"
	"regexp"
	"strings"
)

type ComponentParams struct {
	Name               string
	ConfigurationNames []string
	Vars               map[string]any
}

type refsType map[string]any

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

type componentType struct {
	Configurations map[string]configurationType `validate:"required"`
	Vars           varsType
	Refs           refsType
}

var (
	varsPattern         = regexp.MustCompile(`\$vars\.[^\s]+`)
	refsPattern         = regexp.MustCompile(`^\$refs\.[^\s]+$`)
	yamlPathPattern     = regexp.MustCompile(`^\$((?:\.[^\s.]+)+)?$`)
	dotSeparatedPattern = regexp.MustCompile(`'[^\s]+'|[^.\s]+`)
)

func BuildComponent(source io.Reader, params ComponentParams) (map[string]any, error) {
	var err error
	if params.Name == "" {
		return nil, fmt.Errorf("name param not set")
	}
	component := &componentType{}
	err = parseYamlFile(source, component)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any)
	var configs = params.ConfigurationNames
	if len(configs) == 0 {
		configs = []string{"default"}
	}
	for _, key := range configs {
		configuration, ok := component.Configurations[key]
		if !ok {
			return nil, fmt.Errorf("couldn't find configuration named '%v'", key)
		}
		configRefs := collectRefs(component.Refs, configuration)
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
		configVars, err := collectVars(component, configuration, params)
		if err != nil {
			return nil, err
		}

		err = replacePlaceholdersInMap(body, *varsPattern, configVars)
		if err != nil {
			return nil, err
		}
	}

	return map[string]any{
		params.Name: body,
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
	return ref.(map[string]any), nil
}

func collectRefs(componentRefs refsType, configuration configurationType) refsType {
	var collected = make(refsType)
	if componentRefs != nil {
		collected = deepCopy(map[string]any(componentRefs))
	}
	maps.Copy(collected, configuration.Refs)
	refPrefixedMap := make(refsType, len(collected))
	for k, v := range collected {
		refPrefixedMap["$refs."+k] = v
	}

	return refPrefixedMap
}

func collectVars(component *componentType, configuration configurationType, params ComponentParams) (varsType, error) {
	var collected varsType
	if component.Vars != nil {
		collected = maps.Clone(component.Vars)
	} else {
		collected = make(varsType)
	}
	maps.Copy(collected, configuration.Vars)
	maps.Copy(collected, params.Vars)
	return prependToKeysOfPrimitiveValues(collected, "$vars.")
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
