# Sampling Rules

`Sampling Rules` are defined using [Google Common Expression Language (CEL)](https://github.com/google/cel-spec). The language syntax is defined [here](https://github.com/google/cel-spec/blob/master/doc/langdef.md#syntax). And you can find a list of definitions (operators, functions, and constants) [here](https://github.com/google/cel-spec/blob/master/doc/langdef.md#list-of-standard-definitions) (some functions and complex operators may not be supported).

The `Data Sample` contents are available under the `sample` key scope. This page provides some examples to get you started:

| CEL expression                                 | Description |
|------------------------------------------------|-------------|
| `true`                                         | Matches all `Data Samples`. Useful when you don't know how the data looks like or if you just want to export everything |
| `sample.id != ""`                              | True if the field `id` is defined |
| `sample.id == "some_id" && sample.val == 1`    | You can use boolean expression like `&&` or `||` to concatenate expressions |
| `sample.val1 > 0`                              | Numeric values can be compared |
| `sample.val1 * sample.val2 > 10`               | Arithmetic operations are supported |
| (`sample.val1 || sample.val2) && sample.val3`  | Use parenthesis to specify operator precedence |
