# Examples

## CORS
This example shows you, how to set CORS header for each HTTP request.
```yaml
mokapi: 1.0
workflows:
  - name: CORS
    on:
      http:
    steps:
      - uses: set-response
        with:
          headers: "${{ {'access-control-allow-origin': '*'} }}"
```