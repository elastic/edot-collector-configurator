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

type recipeType struct {
	Args map[string]argsType `validate:"required"`
}

func buildRecipe(source io.Reader, params RecipeParams) ([]byte, error) {
	return nil, fmt.Errorf("unknown error")
}
