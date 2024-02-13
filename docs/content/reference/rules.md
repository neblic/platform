# Rules

`Rules` are defined using [Google Common Expression Language (CEL)](https://github.com/google/cel-spec). The language syntax is defined [here](https://github.com/google/cel-spec/blob/master/doc/langdef.md#syntax) and the the list of supported operators and functions is defined in [here](https://github.com/google/cel-spec/blob/master/doc/langdef.md#list-of-standard-definitions) (some functions and complex operators may not be supported). 

Aditionally, Neblic defines some additional functions that can be used when defining streams or checks (not all functions are supported when creating streams).

!!! note

    The `Data Sample` contents are available under the `sample` key scope. This page provides some examples to get you started:

| Neblic defined function  | Description                                                                                                                                                                                                                          | Stateful | Context             |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------------------- |
| `abs(<v>)`               | Returns the absolut value of `v`. The input can be a double, or an int                                                                                                                                                               | No       | `streams`, `checks` |
| `now()`                  | Returns current time in the collector as a timestamp                                                                                                                                                                                 | No       | `streams`, `checks` |
| `sequence(<v>, <order>)` | Returns true or false depending on if `v` is part of a ordered sequence or not. `order` can be `asc` or `desc`                                                                                                                       | Yes      | `checks`            |
| `complete(<v>, <step>)`  | Returns ture or false depening of in `v` is part of a complete sequence or not. A sequence is complete if all the values are ordered and there is no missing value. `step` contains the distance between each two consecutive values | Yes      | `checks`            |

## Examples

| CEL expression                                             | Description                                                                                               |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `true`                                                     | Matches all `Data Samples`. Useful when creating `Streams` and you just want to select all `Data Samples` |
| `sample.id != ""`                                          | True if the field `id` is defined                                                                         |
| `sample.id == "some_id" && sample.val == 1`                | You can use boolean expression like `&&` or <code>\|\|</code> to concatenate expressions                  |
| `sample.val1 > 0`                                          | Numeric values can be compared                                                                            |
| `sample.val1 * sample.val2 > 10`                           | Arithmetic operations are supported                                                                       |
| <code>(sample.val1 \|\| sample.val2) && sample.val3</code> | Use parenthesis to specify operator precedence                                                            |
| `!sequence(sample.time, "asc")`                            | Matches all the values that are not part of the sequence                                                  |
| `!complete(sample.time, "asc")`                            | Matches all the values that are not part of a complete sequence                                           |