package main

import (
	"io"
)

type RecipeParams struct {
	FeaturesDirPath string
}

type argsType struct {
	Env string
}

type featureDefType struct {
	Source         string `validate:"required"`
	Name           string
	Configurations []string
	Vars           varsType
}

type recipeType struct {
	Args     map[string]argsType       `validate:"required"`
	Features map[string]featureDefType `validate:"required"`
	Services map[string]any            `validate:"required"`
	Const    map[string]any
}

func buildRecipe(source io.Reader, params RecipeParams) ([]byte, error) {
	var err error
	var result []byte
	recipe := &recipeType{}
	err = parseYamlFile(source, recipe)
	if err != nil {
		return nil, err
	}

	return result, nil
}
