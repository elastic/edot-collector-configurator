package main

import (
	"fmt"
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
	return nil, fmt.Errorf("unknown error")
}
