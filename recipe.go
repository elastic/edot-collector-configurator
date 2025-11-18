package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"

	"github.com/goccy/go-yaml"
)

var (
	yamlFileNamePattern = regexp.MustCompile(`(.+)\.[yY][aA]?[mM][lL]`)
	anyArgPattern       = regexp.MustCompile(fmt.Sprintf("%s|%s|%s", `\$const\.[^\s]+`, `\$args\.[^\s]+`, `\$features\.[^\s]+`))
)

type RecipeParams struct {
	Args            map[string]string
	FeaturesDirPath string
}

type argsDefType struct {
	Env string
}

type featureDefType struct {
	Source         string `validate:"required"`
	Name           string
	Configurations []string
	Vars           varsType
}

type recipeType struct {
	Args     map[string]argsDefType    `validate:"required"`
	Features map[string]featureDefType `validate:"required"`
	Services map[string]any            `validate:"required"`
	Const    map[string]any
}

func buildRecipe(source io.Reader, params RecipeParams) ([]byte, error) {
	var err error
	recipe := &recipeType{}
	err = parseYamlFile(source, recipe)
	if err != nil {
		return nil, err
	}
	featureNames, err := getFeatureNames(recipe)
	if err != nil {
		return nil, err
	}
	allArguments, err := collectAllArguments(recipe, params, featureNames)
	if err != nil {
		return nil, err
	}
	builtFeatures := make(map[string]any)
	for k, v := range recipe.Features {
		featureFilePath, err := resolveFeatureFilePath(params, v)
		if err != nil {
			return nil, err
		}
		feature, err := buildFeature(featureNames[k], featureFilePath, v, allArguments)
		if err != nil {
			return nil, err
		}
		mergeMaps(builtFeatures, feature)
	}

	return yaml.Marshal(builtFeatures)
}

func buildFeature(featureName string, featureFilePath string, featureDef featureDefType, arguments map[string]any) (map[string]any, error) {
	vars, err := resolveVars(featureDef.Vars, arguments)
	if err != nil {
		return nil, err
	}
	featureFile, err := os.Open(featureFilePath)
	if err != nil {
		return nil, err
	}
	defer featureFile.Close()

	return BuildFeature(featureFile, FetureParams{
		Name:               featureName,
		ConfigurationNames: featureDef.Configurations,
		Vars:               vars,
	})
}

func collectAllArguments(recipe *recipeType, params RecipeParams, featureNames map[string]string) (map[string]any, error) {
	argsRefs, err := getArgsRefs(recipe.Args, params.Args)
	if err != nil {
		return nil, err
	}
	constRefs, err := getConstantsRefs(recipe.Const)
	if err != nil {
		return nil, err
	}
	featureNameRefs, err := prependToKeysOfPrimitiveValues(featureNames, "$features.")
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
	for k, v := range featureNameRefs {
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

func resolveFeatureFilePath(params RecipeParams, featureDef featureDefType) (string, error) {
	dirPath, err := filepath.Localize(params.FeaturesDirPath)
	if err != nil {
		return "", err
	}
	relativePath, err := filepath.Localize(featureDef.Source)
	if err != nil {
		return "", err
	}
	return filepath.Join(dirPath, relativePath), nil
}

func getFeatureNames(recipe *recipeType) (map[string]string, error) {
	featureNames := make(map[string]string)
	for k, v := range recipe.Features {
		sourceName := filepath.Base(v.Source)
		featureType := yamlFileNamePattern.FindStringSubmatch(sourceName)[1]
		if featureType == "" {
			return nil, fmt.Errorf("could not get feature type from source path: '%s'", v.Source)
		}
		name := featureType
		if len(v.Name) > 0 {
			name = fmt.Sprintf("%s/%s", name, v.Name)
		}
		featureNames[k] = name
	}
	return featureNames, nil
}

func getConstantsRefs(provided map[string]any) (map[string]any, error) {
	return prependToKeysOfPrimitiveValues(provided, "$const.")
}

func getArgsRefs(argsDef map[string]argsDefType, providedArgs map[string]string) (map[string]string, error) {
	collected := maps.Clone(providedArgs)
	var err error
	for k, v := range argsDef {
		_, ok := collected[k]
		if !ok {
			collected[k], err = getEnvVar(v.Env)
			if err != nil {
				return nil, err
			}
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
