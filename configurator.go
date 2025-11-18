package main

import (
	"flag"
	"fmt"
)

var outputFilePath string

func init() {
	flag.StringVar(&outputFilePath, "recipe", "", "path to the YAML recipe file")
	flag.StringVar(&outputFilePath, "output", "otel.yml", "output YAML file path")
}

func main() {
	flag.Parse()

	fmt.Printf("The output is: %s", outputFilePath)
}
