package main

import (
	"flag"
	"fmt"
)

var outputFilePath string

func init() {
	flag.StringVar(&outputFilePath, "output", "otel.yml", "YAML file to write the configuration to")
}

func main() {
	flag.Parse()

	fmt.Printf("The output is: %s", outputFilePath)
}
