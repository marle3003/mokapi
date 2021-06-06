# Actions
Actions are builtin commands which are executed directly in Mokapi instead of an external shell script.

- [read-file](./#read-file-action)
- [mustache](#mustache-action)
- [parse-yaml](#parse-yaml-action)
- [split](#split-action)
- [xpath](#xpath-action)

## read-file action
Read file contents

### Usage
```yaml
steps:
  - uses: read-file
    id: data
    with:
      path: ./data.yml
  - run: echo "${{ steps.data.outputs.content }}"
```

## mustache action
Replace tags in a template. See the [manpage](http://mustache.github.io/mustache.5.html)

### Usage
```yaml
steps:
  - uses: mustache
    id: demo
    with:
      template: 'Say: {{ Label }}'
      data: "${{ [Label: 'Hello World'] }}"
  - run: echo "${{ steps.demo.outputs.result }}"
```

## parse-yaml action
Returns a YAML object for the given input.

### Usage
```yaml
steps:
  - uses: read-file
    id: data
    with:
      path: ./data.yml
  - uses: parse-yaml
    id: demo
    with:
      content: ${{ steps.data.outputs.content }}
  - run: echo "${{ steps.demo.outputs.result }}"
```

## split action
Split a string

### Inputs
- *s*: String to split
- *separator*: The delimiter to split the string.
- *n*: Maximum number of splits. Default: *-1* (no limit)

### Outputs
- *_0*, *_1*,...,*_n*: Each result of a split

### Usage
```yaml
steps:
  - uses: split
    id: demo
    with:
      s: "Hello World"
      separator: " "
  - run: echo "${{ steps.demo.outputs._0 + " " +  steps.demo.outputs._1 }}"
```

## xpath action
Return the string value of the first node matched by the expression

### Usage
```yaml
steps:
  - uses: xpath
    id: demo
    with:
      content: '<book category="cooking"><title lang="en">Everyday Italian</title><author>Giada De Laurentiis</author><year>2005</year><price>30.00</price></book>'
      expression: /bookstore/book/title
  - run: echo "${{ steps.demo.outputs.result }}"
```