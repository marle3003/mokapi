# Actions
Actions are builtin commands which are executed directly in Mokapi instead of an external shell script.

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