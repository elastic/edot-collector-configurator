package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	if len(args) > 3 {
		flagSetArgs = append(flagSetArgs, args[3:]...)
	}
	fs.Parse(flagSetArgs)

	fmt.Printf("The output path is: %s and the recipe: %v", *outputPath, recipe)
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

func getComponentsDirPath() string {
	executable, err := os.Executable()
	checkError(err)
	return filepath.Join(filepath.Dir(executable), "components")
}

func getRecipe(recipeFilePath string) recipeType {
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

func printError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}
