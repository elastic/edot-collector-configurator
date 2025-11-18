package main

import (
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

func parseYamlFile(data io.Reader, result any) error {
	validate := validator.New()
	dec := yaml.NewDecoder(
		data,
		yaml.Validator(validate),
		yaml.Strict(),
	)
	err := dec.Decode(result)
	if err != nil {
		return err
	}
	return nil
}
