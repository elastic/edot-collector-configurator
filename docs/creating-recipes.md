# Creating recipes

## Quickstart

- Create a YAML file to define your recipe. You can place it anywhere — just be sure to pass the appropiate file path to the configurator script later when you use it.
- Add items to the file as needed. The sections and item types you can include are explained below.

For inspiration, you can browse existing recipes provided in the repository.

## Structure Overview

The structure of a recipe consists of descriptive metadata, user-provided arguments, constants, components, and finally the service configuration. Below is an annotated example:

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

Arguments define the values required from the user—these can come from CLI flags or environment variables. Arguments are referenced throughout the recipe using $args.<name>.

```yaml
args: # Provided by the user from either the command line or an environment variable.
  elastic_endpoint: # The name of the argument. Used as command line argument name after "-A", e.g: "-Aelastic_endpoint".
    description: Your Elasticsearch endpoint
    env: ELASTIC_URL # The name of the asociated environment variable for this argument. This will be looked out for when no command line argument is provided.
```

Use arguments whenever you need user-configurable input—for example, endpoints, credentials, or feature toggles.

### Components

Components define the actual building blocks of your recipe — receivers, processors, exporters, connectors, and so on. These components are sourced from the [components](../components/) directory.

> [!NOTE]
> Can't find a component you need? - You can follow [this guide](../docs/creating-components.md) to create one.

Each component can specify:

- The component file (source).
- Which configuration variants to use from the component (configurations).
- An optional name override (name) - If provided, it will be added after the `type` with `[/name], as specified in the upstream OTel collector config.
- Variables (vars) that override defaults inside the component itself.

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

Components can refer to:

- Arguments ($args.<name>)
- Constants ($const.<name>)
- Other component names ($components.<component-name>)

This makes complex configuration generation flexible and reusable.

### Service

The service block follows the same structure as the [upstream OpenTelemetry Collector configuration](https://opentelemetry.io/docs/collector/configuration/#service). The only difference is that instead of writing component names directly, you reference your defined components using $components.<name>.

```yaml
service: 
  pipelines:
    traces:
      receivers: [ $components.my-component-name ] # The component references will be replaced by their final names.
```