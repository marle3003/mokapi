---
title: appendString( path, s )
description: Appends a string to a file at the given path.
---
# appendString( path, s )

Appends the string `s` to a file at the given path.

| Parameter | Type   | Description                  |
|-----------|--------|------------------------------|
| path      | string | Path to the file to write    |
| s         | string | The string content to append |

If the path is relative, Mokapi resolves it relative to the **entry script file**.

If the file does not exist, it will be created. If it exists, the string will be appended.

## Example Appending File

```javascript
import { appendString, writeString, read } from 'mokapi/file'

export default function() {
    writeString('data.json', 'Hello World')
    appendString('data.json', '!')
    console.log(read('data.json'))
}
```