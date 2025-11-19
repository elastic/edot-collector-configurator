package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		printHelpMessage()
		return
	}
	switch args[1] {
	case "build":
		buildRecipe(args[2], args[3:])
	case "info":
		printRecipeInfo(args[2])
	case "help":
		printHelpMessage()
	default:
		fmt.Printf("error: unknown command - %q\n", args[1])
		printHelpMessage()
	}
}

var helpMessage = `
USAGE
  configurator [subcommand]

SUBCOMMANDS
  info   path/to/recipe.yml                       Displays information about the provided recipe and its arguments.
  build  path/to/recipe.yml [-output=file.yml]    Builds a configuration based on the recipe file provided.
`

func printHelpMessage() {
	fmt.Println(helpMessage)
}

func buildRecipe(recipePath string, args []string) {
	recipe := getRecipe(recipePath)
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	fs.String("output", "otel.yml", "Output YAML file path")
	panic("Implement" + fmt.Sprintf("%v", recipe))
}

var infoTemplate = `
Recipe path: %s
Description: %s
Arguments:
%s
`

func printRecipeInfo(recipePath string) {
	recipe := getRecipe(recipePath)
	argsDescription := ""
	for k, v := range recipe.Args {
		argsDescription += fmt.Sprintf("  -A%s - %s", k, v.Description)
		if v.Env != "" {
			argsDescription += fmt.Sprintf(" (ENV var '%s')", v.Env)
		}
		argsDescription += "\n"
	}
	fmt.Printf(infoTemplate, recipePath, recipe.Description, argsDescription)
}

func getComponentsDirPath() string {
	executable, err := os.Executable()
	checkError(err)
	return filepath.Join(filepath.Dir(executable), "components")
}

func getRecipe(recipeFilePath string) recipeType {
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
