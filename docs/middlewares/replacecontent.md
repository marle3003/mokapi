---
navigation: Middlewares/ReplaceContent
---
# ReplaceContent

Replace a string from resources selected by regex with a value defined by replacement.

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