---
title: read( path )
description: Reads the contents of a file and returns it as a string.
---
# read( path )

Reads the file at the given path until EOF and returns its contents.

| Parameter | Type   | Description               |
|-----------|--------|---------------------------|
| path      | string | Path to the file to read  |

If the path is relative, Mokapi resolves it relative to the **entry script file**.

## Returns

| Type   | Description             |
|--------|-------------------------|
| string | The content of the file |

## Example Reading File

```javascript
import { read } from 'mokapi/file'

export default function() {
    const data = read('data.json')
    console.log(data)
}
```