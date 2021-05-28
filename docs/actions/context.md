# Contexts and expressions in Mokapi Actions
With expressions, you can set variables and access contexts. An expression can be any
combination of literal values, references to context, or functions.

Mokapi only evaluates an expression if you use a specific syntax, otherwise it is treated
as a string.

*${{ \<expression\> }}*

In some cases you may omit the expression syntax because Mokapi automatically evaluates it
as an expression. For example *if* conditional.

### Example setting an environment variable
```yaml
env:
  MY_VAR: ${{ 25 + 35 }}
```

## Contexts
Contexts are used to access information about workflow runs or steps.

- **env:** Contains environment variables set in a workflow or step
- **steps:** Information about the steps that have been run in this workflow

You can use the following operators when you access a type member
- . (member access) to access a member of type
- [] (array element) to access an array element

## Literals
As part of an expression you can use *number*, *string* or *boolean* data types.

### Example
```yaml
env:
  myInteger: ${{ 42 }}
  myFloatNumber: ${{ 3.141 }}
  myString: ${{ 'Hello World' }}
  myBoolean: ${{ true }}
```

## Operators
Mokapi Actions provides a number of operators allowing you to perform basic operations
with values. For more information, see [Operators](operators.md)

## Functions
Mokapi offers a set of builtin functions which you can use in expressions.