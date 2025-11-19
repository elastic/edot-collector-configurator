package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var recipeFilePath string
var outputFilePath string

func init() {
	flag.StringVar(&recipeFilePath, "recipe", "", "Path to the YAML recipe file")
	flag.StringVar(&outputFilePath, "output", "otel.yml", "Output YAML file path")
}

func main() {
	flag.Parse()
	recipe := getRecipe()
	fmt.Printf("The components dir is: %s\n", getComponentsDirPath())
	fmt.Printf("The recipe is: %v", recipe)
}

func getComponentsDirPath() string {
	executable, err := os.Executable()
	checkError(err)
	return filepath.Join(filepath.Dir(executable), "components")
}

func getRecipe() recipeType {
	if recipeFilePath == "" {
		panic(fmt.Errorf("recipe path not provided - pass it using the '-recipe' argument"))
	}
	wd, err := os.Getwd()
	checkError(err)
	f, err := os.Open(filepath.Join(wd, recipeFilePath))
	checkError(err)
	defer f.Close()

	recipe, err := ParseRecipe(f)
	checkError(err)
	return recipe
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
