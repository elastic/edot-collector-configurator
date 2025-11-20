# Creating recipes

## Quickstart

- Create a yaml file for your recipe. It could be place anywhere, just make sure to pass the full path to it when calling the configurator script.
- Add items to it depdending on your needs. You can find the types of items you can add below.

You can take a look at some existing recipes for inspiration.

## Structure Overview

```yaml
description: "Explains the outcome of this recipe"

args: # Provided by the user from either the command line or an environment variable.
  elastic_endpoint: # The name of the argument. Used as command line argument name after "-A", e.g: "-Aelastic_endpoint".
    description: Your Elasticsearch endpoint
    env: ELASTIC_URL # The name of the asociated environment variable for this argument. This will be looked out for when no command line argument is provided.

const: # This is just handy to avoid duplicating values across this file.
  some_constant: some value

components: # Set of components that form this recipe.

  my-otlp: # This name won't be printed in the final config, is only for internal recipe reference purposes.
    source: receivers/otlp.yml # The path to the recipe. Must be relative to the root of this repo.
    configurations: [ http, grpc ] # The set of configurations needed from this repo. If no configurations are provided, a configuration named "default" from the component will be used.
    vars: # These are referenced inside the component file and will override any default vars from there.
      http_port: 4318

  some-exporter:
    source: exporters/elasticsearch.yml
    name: custom-name # This name is optional. When provided will be appended to the type like so: "type/name". This is based on naming configs from the uptream specs.
    vars:
      elastic_endpoint: $args.elastic_endpoint 
  elasticapm-connector:

service: # The same upstream structure: https://opentelemetry.io/docs/collector/configuration/#service
  pipelines:
    traces:
      receivers: [ $components.my-otlp ] # The component references will be replaced by their final names.
      exporters: [ $components.some-exporter ]
```

### Args

The args contain a map of arguments that must be provided either via command line arguments or environment variables.

```yaml
args: # Provided by the user from either the command line or an environment variable.
  elastic_endpoint: # The name of the argument. Used as command line argument name after "-A", e.g: "-Aelastic_endpoint".
    description: Your Elasticsearch endpoint
    env: ELASTIC_URL # The name of the asociated environment variable for this argument. This will be looked out for when no command line argument is provided.
```

### Components

A map of the components that will be used by the recipe. You must search for the components you need from within the [components](../components/) folder.

> [!NOTE]
> Can't find a component you need? - You can follow [this guide](../docs/creating-components.md) to create one.

```yaml
components: # Set of components that form this recipe.
  my-otlp: # This name won't be printed in the final config, is only for reference purposes within this the recipe.
    source: receivers/otlp.yml # The path to the recipe. Must be relative to the root of this repo.
    configurations: [ http, grpc ] # [OPTIONAL] The set of configurations needed from this repo. If no configurations are provided, a configuration named "default" from the component will be used.
    name: custom-name # [OPTIONAL] When provided, it will be appended to the type like so: "type/name". This is based on naming configs from the uptream specs. 
    vars: # These are referenced inside the component file and will override any default vars from there.
      http_port: 4318
      endpoint: $args.some_arg and $const.some_constant and $components.other-component # You can add args, const and other component names here.
```

### Service

This is the exact same `service` config defined upstream: https://opentelemetry.io/docs/collector/configuration/#service

The only difference is that we'll use references to components as its values.

```yaml
service: 
  pipelines:
    traces:
      receivers: [ $components.my-component-name ] # The component references will be replaced by their final names.
```