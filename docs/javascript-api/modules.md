---
title: JavaScript Modules
description: Learn how to organize Mokapi scripts with built-in, local, and JSON/YAML modules for better maintainability and flexibility.
---
# JavaScript Modules

When writing scripts, it is common to split code  into multiple files or to use third-party modules.
Mokapi supports importing three types of modules:

- **Built-in modules**: Provided by Mokapi for various functionalities.
- **Local filesystem modules**: Custom scripts and Node.js packages.
- **JSON & YAML modules**: Configuration files converted into JavaScript objects.

``` box=tip
Mokapi monitors all imported modules with `fsnotify`. If a module is modified, any dependent script is automatically reloaded.
```

## Built-in Modules

Mokapi offers built-in modules like `faker`, which generates realistic test data from JSON schemas. 
See the [API documentation](/docs/javascript-api/overview.md) for a complete list.

```javascript
import { fake } from 'mokapi/faker'
```

## Local Filesystem Modules

You can import files using relative or absolute paths, and Mokapi supports Node.js modules.

```javascript
import { someFunc } from './helpers.js'
import { otherFunc } from '../lib'
import dateTime from 'date-time' // Requires: npm install date-time
```

## JSON & YAML Modules

Mokapi allows importing JSON and YAML files, automatically converting them into JavaScript objects.

```javascript tab=Javascript
import users from './users.json'
import envs from './environments.yaml'

console.log(users[0].name)
console.log(envs[0])
```
```json tab=JSON
[  
    { "name":"Alice", "email":"alice@foo.bar" },  
    { "name":"Bob", "email":"bob@foo.bar" }  
]  
```
```yaml tab=YAML
- test
- integration
- production
```

By leveraging these module types, you can create flexible, maintainable, and scalable Mokapi scripts.