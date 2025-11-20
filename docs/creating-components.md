# Creating components

Components are reusable parts/fragments of an EDOT Collector config file that are referenced in recipes.

## Structure

This is the full structure. More details are explained below.

> [!NOTE]
> Only the `configurations.[name].content` object is mandatory. The rest are optional.

```yaml
vars: # [OPTIONAL] Map with "global vars". These can be overridden per configuration or from the recipe.
  test-var: global value
  test-var2: global value 2
refs: {} # [OPTIONAL] References to map structures that are shared across configurations. Useful to avoid repeating map keys/values.
configurations: # Defines different "variants" for this component.
  [config-name]: # Used later from a recipe to choose the config variant it needs.
    content: # The component's content.
      some_key: some value
      another_key: a value that uses a var, $vars.test-var.
    refs: {} # [OPTIONAL] References to a map that's specific to this config. More info below.
    vars: # [OPTIONAL] Map with "config-specific" vars, which can be overridden from the recipe.
      test-var2: This value overrides the "global var" only for this config.
    append: # [OPTIONAL] Appends content to either maps or lists. Useful when the main content comes from a global ref.
      - path: "$.some.key" # yaml path to the target
        content: {} # content to add to the target specified in the path
```

### Configurations

A component can have multiple configurations, for example, this is the upstream's [otlp config](https://opentelemetry.io/docs/collector/configuration/#receivers) definition:

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

In this case, the `otlp` receiver component has a `grpc` and an `http` variant.

We can translate that example `otlp` config to a configurator component like this:

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

> [!INFORMATION]
> The `content` only has the `protocols` item, and it doesn't add its parents (`otlp` and `receivers`). This is intentional, as those parents will be added later during the recipe build.

You may define as many configurations as needed. Make sure to select the ones you want for a recipe within the recipe file.

### Vars

You can define variables using the `vars` key either at the root of the component (to define global vars) or at each configuration's definition.

Variables can be provided from the recipe file, in which case the values provided there will override any values defined inside the component file.

> [!IMPORTANT]
> Vars can only contain primitive values (string, boolean and numbers). The configurator will raise an error if an object is set there.

```yaml
vars: # Map with "global vars". These can be overridden per configuration or from the recipe.
  test-var: global value
  test-var2: global value 2
configurations: 
  my-config-nam: 
    content: 
      some_key: some value $vars.test-var2
      another_key: a value that uses a var, $vars.test-var.
    vars: 
      test-var2: This value overrides the global var "test-var2" defined earlier only for this config's content.
```

### Refs

Refs are references to maps so that they can be embedded in other maps, which helps to avoid repeating common map structures across different configurations.

In our `otlp` component example from [above](#configurations), we had the following configurations:

```yaml
# Without using refs
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

In this scenario, both configurations share the same root key `protocols`. If we wanted to define it only once and reuse it for every config,
we can do so this way:

```yaml
# Reusing common structures with refs
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
    content:
    refs:
      protocol_details:
        http:
          endpoint: 0.0.0.0:4318
```

During the recipe build, the refs are resolved and merged on each configuration that uses them.

### Append

When defining base map structures using [refs](#refs), sometimes the base structure misses some extra keys that are needed for a specific configuration only. Append helps adding those items per configuration.

Using our previous example [refs](#refs), we can add a new key at the same level as `protocols` like this:

```yaml
# Adding configuration-specific items with append
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
    content:
    refs:
      protocol_details:
        http:
          endpoint: 0.0.0.0:4318
    append:
      - path: "$" # This value must be a "YAML Path". The "$" represents the root object of an item in a YAML Path.
        content:
          something: some extra value
```

The final output of building the `http` configuration would look like the following:

```yaml
# Final form of the http configuration after used in a recipe
otlp:
  protocols:
    http:
      endpoint: 0.0.0.0:4318
    something: some extra value
```

> [!NOTE]
> The `path` object used in an `append` item uses a YAML path format. (well, partially supported format, you can only create paths that lead to maps or lists).