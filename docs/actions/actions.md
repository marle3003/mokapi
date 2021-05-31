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