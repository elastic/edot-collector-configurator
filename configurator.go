package main

import (
	"flag"
	"fmt"
	"os"
)

var recipeFilePath string
var outputFilePath string

func init() {
	flag.StringVar(&recipeFilePath, "recipe", "", "Path to the YAML recipe file")
	flag.StringVar(&outputFilePath, "output", "otel.yml", "Output YAML file path")
}

func main() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Ex: %s", executable)

	flag.Parse()
}
