# Rules

*Rules* are used to define *Streams* and to generate *Events*. Both share the same syntax altough they have different function sets available. They are defined using the [Google Common Expression Language (CEL)](https://github.com/google/cel-spec). 

The language syntax is defined [here](https://github.com/google/cel-spec/blob/master/doc/langdef.md#syntax) and the the list of supported operators and functions common to *Streams* and *Events* is defined in [here](https://github.com/google/cel-spec/blob/master/doc/langdef.md#list-of-standard-definitions). The table below shows the additional functions that Neblic defines:

| Neblic defined function  | Description                                                                                                                                                                                                                          | Stateful | Context             | Keyed |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------------------- | ----- |
| `abs(<v>)`               | Returns the absolut value of *v*. The input can be a double, or an int                                                                                                                                                               | No       | *streams*, *events* | false |
| `now()`                  | Returns current time in the collector as a timestamp                                                                                                                                                                                 | No       | *streams*, *events* | false |
| `sequence(<v>, <order>)` | Returns true or false depending on if *v* is part of a ordered sequence or not. *order* can be *asc* or *desc*                                                                                                                       | Yes      | *events*            | true  |
| `complete(<v>, <step>)`  | Returns ture or false depening of in *v* is part of a complete sequence or not. A sequence is complete if all the values are ordered and there is no missing value. *step* contains the distance between each two consecutive values | Yes      | *events*            | true  |

## Examples

| CEL expression                                             | Description                                                                                               |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `true`                                                     | Matches all *Data Samples*. Useful when creating *Streams* and you just want to select all *Data Samples* |
| `sample.id != ""`                                          | True if the field *id* is defined                                                                         |
| `sample.id == "some_id" && sample.val == 1`                | You can use boolean expression like `&&` or <code>\|\|</code> to concatenate expressions                  |
| `sample.val1 > 0`                                          | Numeric values can be compared                                                                            |
| `sample.val1 * sample.val2 > 10`                           | Arithmetic operations are supported                                                                       |
| <code>(sample.val1 \|\| sample.val2) && sample.val3</code> | Use parenthesis to specify operator precedence                                                            |
| `!sequence(sample.time, "asc")`                            | Matches all the values that are not part of the sequence                                                  |
| `!complete(sample.time, "asc")`                            | Matches all the values that are not part of a complete sequence                                           |
