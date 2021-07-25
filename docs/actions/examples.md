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

## Filtering
If you have a */models* endpoint which are models of motorbikes, you can filter via the
property name such as *GET /models?make=bmw*. For this example create a file bikes.yml 
with the following content:
```yaml
bikes:
  - name: R 1250 GS
    make: BMW
  - name: R 1250 RT
    make: BMW
  - name: Africa Twin
    make: Honda
  - name: Versys 1000 SE
    make: Kawasaki
```
The following workflow loads the file and parses the content. The last step filters the
bikes by the request parameter.
```yaml
mokapi: 1.0
workflows:
  - name: filter bikes
    on:
      http:
        get: /models
    steps:
      - uses: read-file
        id: file
        with:
          path: ./bikes.yml
      - uses: parse-yaml
        id: parse
        with:
          content: ${{ steps.file.outputs.content }}
      - uses: set-response
        if: request.query.make != null
        with:
          data: ${{ findAll(steps.parse.outputs.result.bikes,
            x => toLower(x.make) == toLower(request.query.make)) }}
      - uses: set-response
        if: request.query.make == null
        with:
          data: ${{ steps.parse.outputs.result.bikes }}
```
### Using query parameter for any property
If your endpoint has a query parameter *q* to search any models which contains the value
of *q* in any property, you can solve this by the following statement.
```
${{ findAll(steps.parse.outputs.result.bikes, 
    x => any(x.*, y => contains(toLower(y), toLower(request.query.q)))) }}
```

## Pagination
Following example shows you, how to mock a simple pagination endpoint. A file provides an
individual page, and the name contains the associated offset.

```yaml
mokapi: 1.0
workflows:
  - name: Items
    on:
      http:
        get: /items
    steps:
      - uses: read-file
        id: file
        with:
          path: ${{ format("items{0}.txt", request.query.offset) }}
      - uses: set-response
        with:
         body: ${{ steps.file.outputs.content }}
```