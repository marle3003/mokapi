# ReplaceContent

Filters your content from your resources. The ReplaceContent middleware only replaces string resources and only supports request body.

## Request Body

```yaml
# replaces regex ___QueryId___ with the value selected by XPath //Method[@Name='ExecuteQuery']/@Id
x-mokapi-middlewares:
  - type: replaceContent
    replacement:
      from: requestBody
      selector: //Method[@Name='ExecuteQuery']/@Id
    regex: ___QueryId___
```