# Creating Components

Components are reusable fragments of an EDOT Collector configuration.
They encapsulate pieces of YAML that can be parameterized, reused, and
referenced within recipes.

## Structure Overview

Below is the full component structure. Only
`configurations.[name].content` is required; all other fields are
optional.

```yaml
vars:
  test-var: global value
  test-var2: global value 2

refs: {}

configurations:
  [config-name]:
    content:
      some_key: some value
      another_key: a value using a var, $vars.test-var.

    refs: {}

    vars:
      test-var2: Overrides the global var for this configuration.

    append:
      - path: "$.some.key"
        content: {}
```

## Configurations

A component can have multiple configurations, and each one must define its contents within a `content` object. 

```yaml
# Simple configuration example
configurations:
  my-config-name:
    content:
      my-key: my value
```

### Example

For example, this is the upstream's [otlp config](https://opentelemetry.io/docs/collector/configuration/#receivers) definition:

```yaml
# Extract of an upstream config found here: https://opentelemetry.io/docs/collector/configuration/#receivers
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
```

We can translate that example `otlp` config to a component, like this:

```yaml
configurations:
  grpc:
    content:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317

  http:
    content:
      protocols:
        http:
          endpoint: 0.0.0.0:4318
```

When a single configuration is defined for a component, it should be named `default`. Configurations named `default` are used by the
recipes that do not specify any configuration name to use from a component file.

## Vars

Variables may be declared globally or per configuration and they can be overridden by variables defined from the recipe file.

You can reference them within any content's value (even content from [refs](#refs) and [append](#append) blocks) using the `$vars.` prefix, as shown in the example below.

> [!IMPORTANT]
> Vars can only contain primitive values (string, boolean and numbers). The configurator will raise an error if an object is set there.

```yaml
vars:
  test-var: global value
  test-var2: global value 2

configurations:
  my-config-name:
    content:
      some_key: some value $vars.test-var2
      another_key: a value using $vars.test-var
    vars:
      test-var2: Overrides the global value with the same name (test-var2)
```

## Refs

Refs are references to maps that can be embedded in other maps, which helps to avoid repeating common map structures across different configurations.

``` yaml
# Without refs:
configurations:
  grpc:
    content:
      protocols: # Repeated key across configs
        grpc:
          endpoint: 0.0.0.0:4317
  http:
    content:
      protocols: # Repeated key across configs
        http:
          endpoint: 0.0.0.0:4318
```

```yaml
# With refs:
refs:
  base:
    protocols: $refs.protocol_details # This ref will be defined in each configuration

configurations:
  grpc:
    content: $refs.base
    refs:
      protocol_details:
        grpc:
          endpoint: 0.0.0.0:4317

  http:
    content: $refs.base
    refs:
      protocol_details:
        http:
          endpoint: 0.0.0.0:4318
```

During the recipe build, the refs are resolved and merged on each configuration that uses them.

## Append

When defining base map structures using [refs](#refs), sometimes the base structure misses some extra keys that are needed for a specific configuration only. Append helps adding those items per configuration.

Adding config‑specific items:

```yaml
refs:
  base:
    protocols: $refs.protocol_details

configurations:
  grpc:
    content: $refs.base
    refs:
      protocol_details:
        grpc:
          endpoint: 0.0.0.0:4317

  http:
    content: $refs.base
    refs:
      protocol_details:
        http:
          endpoint: 0.0.0.0:4318
    append:
      - path: "$" # This value must be a "YAML Path". The "$" represents the root object of an item in a YAML Path.
        content: # The content can be either a map or a list (depending on the target defined in the path).
          something: some extra value
```

The final output of building the `http` configuration would look like the following:

```yaml
# Resolved result
otlp:
  protocols:
    http:
      endpoint: 0.0.0.0:4318
  something: some extra value
```

> [!NOTE]
> The `path` object used in an `append` item uses a YAML path format. Only simple paths to maps or lists are supported.

## Location of the component file

Components MUST be located within the [components](../components) folder and under the directory that fits its category.

For example, if we wanted to create a processor component named `debug`, we must locate it in a file named `debug.yml` within the `components/processors` folder, like so:

```
components/
├─ processors/
│  ├─ debug.yml # This will be our new component file
```

This is needed because the configurator script will take the file name (without the .yml part) as the component's type, and its parent folder name as the component's category.

Considering that this is our component file contents:

```yaml
# Contents of components/processors/debug.yml
configurations:
  default:
    something: some value
```

This is how that component will be added to the final config:

```yaml
processors:
  debug: # If a custom name is provided, e.g. "my-name", then the final result will be: "debug/my-name".
    something: some value
```