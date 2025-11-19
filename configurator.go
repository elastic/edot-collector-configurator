package main

import (
	"flag"
)

var recipeFilePath string
var outputFilePath string

func init() {
	flag.StringVar(&recipeFilePath, "recipe", "", "Path to the YAML recipe file")
	flag.StringVar(&outputFilePath, "output", "otel.yml", "Output YAML file path")
}

func main() {
	flag.Parse()
	// fs := flag.NewFlagSet("args", flag.ContinueOnError)

}
