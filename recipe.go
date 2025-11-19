package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
)

var (
	yamlFileNamePattern = regexp.MustCompile(`(.+)\.[yY][aA]?[mM][lL]`)
	anyArgPattern       = regexp.MustCompile(fmt.Sprintf("%s|%s|%s", `\$const\.[^\s]+`, `\$args\.[^\s]+`, `\$components\.[^\s]+`))
)

type RecipeParams struct {
	Args              map[string]string
	ComponentsDirPath string
}

type argsDefType struct {
	Description string `validate:"required"`
	Env         string
}

type componentDefType struct {
	Source         string `validate:"required"`
	Name           string
	Configurations []string
	Vars           varsType
}

type recipeType struct {
	Args        map[string]argsDefType      `validate:"required"`
	Description string                      `validate:"required"`
	Components  map[string]componentDefType `validate:"required"`
	Services    map[string]any              `validate:"required"`
	Const       map[string]any
}

func ParseRecipe(source io.Reader) (recipeType, error) {
	recipe := &recipeType{}
	err := parseYamlFile(source, recipe)
	return *recipe, err
}

func BuildRecipe(recipe *recipeType, params RecipeParams) (map[string]any, error) {
	var err error
	componentNames, err := getComponentNames(recipe)
	if err != nil {
		return nil, err
	}
	allArguments, err := collectAllArguments(recipe, params, componentNames)
	if err != nil {
		return nil, err
	}
	builtComponents := make(map[string]any)
	for k, v := range recipe.Components {
		componentFilePath := filepath.Join(params.ComponentsDirPath, v.Source)
		component, err := buildComponent(componentNames[k], componentFilePath, v, allArguments)
		if err != nil {
			return nil, err
		}
		componentName := filepath.Base(filepath.Dir(componentFilePath))
		err = mergeMaps(builtComponents, map[string]any{
			componentName: component,
		})
		if err != nil {
			return nil, err
		}
	}
	resolvedServices := maps.Clone(recipe.Services)
	err = replacePlaceholdersInMap(resolvedServices, *anyArgPattern, allArguments)
	if err != nil {
		return nil, err
	}
	err = mergeMaps(builtComponents, map[string]any{
		"services": resolvedServices,
	})
	if err != nil {
		return nil, err
	}

	return builtComponents, nil
}

func buildComponent(componentName string, componentFilePath string, componentDef componentDefType, arguments map[string]any) (map[string]any, error) {
	vars, err := resolveVars(componentDef.Vars, arguments)
	if err != nil {
		return nil, err
	}
	componentFile, err := os.Open(componentFilePath)
	if err != nil {
		return nil, err
	}
	defer componentFile.Close()

	return BuildComponent(componentFile, ComponentParams{
		Name:               componentName,
		ConfigurationNames: componentDef.Configurations,
		Vars:               vars,
	})
}

func collectAllArguments(recipe *recipeType, params RecipeParams, componentNames map[string]string) (map[string]any, error) {
	argsRefs, err := getArgsRefs(recipe.Args, params.Args)
	if err != nil {
		return nil, err
	}
	constRefs, err := getConstantsRefs(recipe.Const)
	if err != nil {
		return nil, err
	}
	componentNameRefs, err := prependToKeysOfPrimitiveValues(componentNames, "$components.")
	if err != nil {
		return nil, err
	}
	allValues := make(map[string]any)
	for k, v := range argsRefs {
		allValues[k] = v
	}
	for k, v := range constRefs {
		allValues[k] = v
	}
	for k, v := range componentNameRefs {
		allValues[k] = v
	}
	return allValues, nil
}

func resolveVars(varsType varsType, arguments map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	for k, v := range varsType {
		if isString(v) {
			resolved, err := resolvePlaceholdersInString(v.(string), *anyArgPattern, arguments)
			if err != nil {
				return nil, err
			}
			result[k] = resolved
		} else {
			result[k] = v
		}
	}
	return result, nil
}

func getComponentNames(recipe *recipeType) (map[string]string, error) {
	componentNames := make(map[string]string)
	for k, v := range recipe.Components {
		sourceName := filepath.Base(v.Source)
		componentType := yamlFileNamePattern.FindStringSubmatch(sourceName)[1]
		if componentType == "" {
			return nil, fmt.Errorf("could not get component type from source path: '%s'", v.Source)
		}
		name := componentType
		if len(v.Name) > 0 {
			name = fmt.Sprintf("%s/%s", name, v.Name)
		}
		componentNames[k] = name
	}
	return componentNames, nil
}

func getConstantsRefs(provided map[string]any) (map[string]any, error) {
	return prependToKeysOfPrimitiveValues(provided, "$const.")
}

func getArgsRefs(argsDef map[string]argsDefType, providedArgs map[string]string) (map[string]string, error) {
	var collected map[string]string
	if providedArgs != nil {
		collected = maps.Clone(providedArgs)
	} else {
		collected = make(map[string]string)
	}
	for k, v := range argsDef {
		_, ok := collected[k]
		if !ok {
			envVarValue, err := getEnvVar(v.Env)
			if err != nil {
				return nil, fmt.Errorf("arg '%s' not provided - you may provide via the env var: '%s' or via the command line argument: '-A%s'", k, v.Env, k)
			}
			collected[k] = envVarValue
		}
	}
	return prependToKeysOfPrimitiveValues(collected, "$args.")
}

func getEnvVar(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("env var not found: '%s'", name)
	}
	return value, nil
}
