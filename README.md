# EDOT Collector Configurator

Creates configuration files for the [EDOT Collector](https://www.elastic.co/docs/reference/edot-collector) based on predefined use cases, or "recipes".

# Usage

## 💻 System requirements

- [Golang](https://go.dev/doc/install) 

## 📄 Building a configuration

### Step 1 - look for a recipe

Look for a recipe that you'd like to use within the [recipes](recipes) dir.

### Step 2 - learn about the recipe's arguments

Recipes might require arguments to be provided wither via the command line or via environment variables, you can learn
about those arguments and also get a detaild description of what the recipe does, by running the following command:

```shell
./configurator info path/to/recipe.yml
```

### Step 3 (final step) - build your configuration
Build your config file using the previously chosen recipe

```shell
./configurator build path/to/recipe.yml [-output=otel.yml] [recipe args as shown in step 2]
```

### Example

Let's use this test recipe: [recipes/gateway/test/otlp.yml](recipes/gateway/test/otlp.yml)

First, we learn about its args:

```shell
./configurator info recipes/gateway/test/otlp.yml
```
```txt
## This is the output of the info command

DESCRIPTION
  Receives OTLP data over HTTP (on port 4318) and gRPC (on port 4317) and exports it to Elasticsearch.

ARGUMENTS
  -Aelastic_endpoint   Your Elasticsearch endpoint (ENV var 'ELASTIC_URL')
  -Aelastic_api_key    Your Elasticsearch API Key (ENV var 'ELASTIC_API_KEY')
```

Then we can **build** a config from it like so:

```shell
# Optional, setting one of the ARGs via environment variable
export ELASTIC_API_KEY=MY_ES_API_KEY
./configurator build recipes/gateway/test/otlp.yml -Aelastic_endpoint=http://localhost:9200
```

After that, we should find a file named `otel.yml` created in the working dir with our final config file.

# License

This software is licensed under the [Apache 2](LICENSE) license.