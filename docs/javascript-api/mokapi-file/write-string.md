---
title: writeString( path, s )
description: Writes a string to a file at the given path.
---
# writeString( path, s )

Writes the string `s` to a file at the given path.

| Parameter | Type   | Description                   |
|-----------|--------|-------------------------------|
| path      | string | Path to the file to write     |
| s         | string | The string content to write   |

If the path is relative, Mokapi resolves it relative to the **entry script file**.

If the file does not exist, it will be created. If it exists, it will be overwritten.

## Example Writing File

```javascript
import { writeString, read } from 'mokapi/file'

export default function() {
    writeString('data.json', 'Hello World')
    console.log(read('data.json'))
}
```