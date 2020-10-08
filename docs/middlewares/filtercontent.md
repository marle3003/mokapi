---
navigation: Middlewares/FilterContent
---
# FilterContent

Filters your content from your resources

## Query Parameter

```yaml
# filter the content based on the userid that is provided by the query parameter id
x-mokapi-middlewares:
  - type: filterContent
    filter: userId = param["id"]
```

## Request Body

```yaml
# filter the content based on the userid that is provided by the request body parameter id
x-mokapi-middlewares:
  - type: filterContent
    filter: userId = body["id"]
```