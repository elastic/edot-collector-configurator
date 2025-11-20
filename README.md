# EDOT Collector Configurator

The **EDOT Collector Configurator** is a small utility that generates
configuration files for the [EDOT
Collector](https://www.elastic.co/docs/reference/edot-collector) based
on predefined use cases, called **recipes**.\
Recipes allow you to quickly build valid, parameterized collector
configurations without editing YAML manually.

## 💻 System Requirements

-   [Golang](https://go.dev/doc/install)

## 🚀 Quick Start

To generate a configuration:

1.  Pick a recipe from the `recipes/` directory.
2.  Inspect the recipe and view its required arguments.
3.  Build a configuration file using those arguments.

The following sections explain each step in detail.

## 📄 Building a Configuration

### Step 1 - Choose a recipe

Browse the available recipes in the [`recipes`](recipes) directory and
select one that matches your use case.

### Step 2 - View recipe arguments

Recipes may require arguments, which can be provided either on the
command line or via environment variables.

Run:

``` shell
./configurator info path/to/recipe.yml
```

This command prints:

-   A detailed description of what the recipe does
-   A list of required and optional arguments
-   The associated environment variable names (if applicable)

### Step 3 - Build the configuration

Use the recipe and provide the required arguments:

``` shell
./configurator build path/to/recipe.yml [-output=otel.yml] [recipe args...]
```

If `-output` is omitted, the output file defaults to `otel.yml`.

## 🧪 Example

We'll use the test recipe:\
`recipes/gateway/test/otlp.yml`

### 1. Inspect the recipe

``` shell
./configurator info recipes/gateway/test/otlp.yml
```

Example output:

``` txt
DESCRIPTION
  Receives OTLP data over HTTP (on port 4318) and gRPC (on port 4317) and exports it to Elasticsearch.

ARGUMENTS
  -Aelastic_endpoint   Your Elasticsearch endpoint (ENV var 'ELASTIC_URL')
  -Aelastic_api_key    Your Elasticsearch API Key (ENV var 'ELASTIC_API_KEY')
```

### 2. Build the configuration

You can pass arguments as flags or environment variables:

``` shell
# Optional: set one argument via an environment variable
export ELASTIC_API_KEY=MY_ES_API_KEY

# Build the configuration file
./configurator build recipes/gateway/test/otlp.yml -Aelastic_endpoint=http://localhost:9200
```

A file named `otel.yml` will be created in the working directory
containing the generated EDOT Collector configuration.

## License

This software is licensed under the [Apache 2](LICENSE) license.
