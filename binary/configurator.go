package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		printHelpMessage()
		return
	}
	switch args[1] {
	case "build":
		buildRecipe(args)
	case "info":
		printRecipeInfo(args)
	case "help":
		printHelpMessage()
	default:
		printError(fmt.Errorf("unknown command - %q", args[1]))
		printHelpMessage()
	}
}

var helpMessage = `
USAGE
  configurator [subcommand]

SUBCOMMANDS
  info   path/to/recipe.yml                       Displays information about the provided recipe and its arguments.
  build  path/to/recipe.yml [-output=otel.yml]    Builds a configuration based on the recipe file provided.
`

func printHelpMessage() {
	fmt.Println(helpMessage)
}

func buildRecipe(args []string) {
	err := checkRecipeProvided(args)
	if err != nil {
		printError(err)
		return
	}
	recipe := getRecipe(args[2])

	fs := flag.NewFlagSet("build", flag.ExitOnError)
	outputPath := fs.String("output", "otel.yml", "Output YAML file path")

	flagSetArgs := []string{}
	recipeArgs := make(map[string]string)
	for k, v := range recipe.Args {
		fs.Func("A"+k, v.Description, func(s string) error {
			recipeArgs[k] = s
			return nil
		})
	}
	if len(args) > 3 {
		flagSetArgs = append(flagSetArgs, args[3:]...)
	}
	fs.Parse(flagSetArgs)

	configuration, err := BuildRecipe(&recipe, RecipeParams{
		Args:              recipeArgs,
		ComponentsDirPath: getComponentsDirPath(),
	})

	checkUnexpectedError(err)
	saveConfiguration(configuration, *outputPath)
}

func saveConfiguration(configuration map[string]any, outputPath string) {
	yamlData, err := yaml.Marshal(configuration)
	checkUnexpectedError(err)
	f, err := os.Create(outputPath)
	checkUnexpectedError(err)
	defer f.Close()

	_, err = f.Write(yamlData)
	checkUnexpectedError(err)
}

func getComponentsDirPath() string {
	executable, err := os.Executable()
	checkUnexpectedError(err)
	return filepath.Join(filepath.Dir(executable), "components")
}

var infoTemplate = `
Recipe path: %s
Description: %s
Arguments:
%s
`

func printRecipeInfo(args []string) {
	err := checkRecipeProvided(args)
	if err != nil {
		printError(err)
		return
	}
	recipePath := args[2]
	recipe := getRecipe(recipePath)
	argsDescription := ""
	longestArgName := 0
	for k := range recipe.Args {
		if len(k) > longestArgName {
			longestArgName = len(k)
		}
	}
	for k, v := range recipe.Args {
		argName := "-A" + k
		argsDescription += fmt.Sprintf("  %s%s%s", argName, strings.Repeat(" ", longestArgName-len(argName)+5), v.Description)
		if v.Env != "" {
			argsDescription += fmt.Sprintf(" (ENV var '%s')", v.Env)
		}
		argsDescription += "\n"
	}
	fmt.Printf(infoTemplate, recipePath, recipe.Description, argsDescription)
}

func checkRecipeProvided(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("you must provide the recipe file name")
	}
	return nil
}

func getRecipe(recipeFilePath string) recipeType {
	wd, err := os.Getwd()
	checkUnexpectedError(err)
	f, err := os.Open(filepath.Join(wd, recipeFilePath))
	checkUnexpectedError(err)
	defer f.Close()

	recipe, err := ParseRecipe(f)
	checkUnexpectedError(err)
	return recipe
}

func checkUnexpectedError(err error) {
	if err != nil {
		panic(err)
	}
}

func printError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}
