# Creating components

Components are reusable parts/fragments of an EDOT Collector config file that are referenced in recipes.

## Structure

> [!NOTE]
> Only the `configurations` object is mandatory. The rest are optional.

```diff
#vars: # Map with "global vars". These can be overridden per configuration or from the recipe.
#  test-var: global value
#  test-var2: global value 2
#refs: {} # References to map structures that are shared across configurations. Useful to avoid repeating map keys/values.
+ configurations: # Defines different "variants" for this component.
+  [config-name]: # Used later from a recipe to choose the config variant it needs.
+    content: # The component's content.
#     some_key: some value
#       another_key: a value that uses a var, $vars.test-var.
#   refs: {} # References to a map that's specific to this config. More info below.
#   vars: # Map with "config-specific" vars, which can be overridden from the recipe.
#     test-var2: This value overrides the "global var" only for this config.
#   append: # Appends content to either maps or lists. Useful when the main content comes from a global ref.
#     - path: "$.some.key" # yaml path to the target
#       content: {} # content to add to the target specified in the path
```