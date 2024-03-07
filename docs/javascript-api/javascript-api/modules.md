---
title: JavaScript Modules
description: When writing scripts, it is common to separate code in different files or to use third-party modules. In Mokapi Scripts you can import three different kinds of modules.
---
# Modules

When writing scripts, it is common to separate code in
different files or to use third-party modules. In Mokapi Scripts you
can import three different kinds of modules:

- Built-in modules
- Local filesystem modules
- JSON module

``` box=tip
Mokapi watches all imported modules for changes using fsnotify.
If the script S imports the module A and the module A is
changed, the script S is also reloaded.
```

## Built-in modules

These modules are provided by Mokapi. For example the `faker` module
used to generate random data for a given JSON schema. For 
a full list, see [the API documentation](/docs/javascript-api/javascript-api/overview.md).

```javascript
import { fake } from 'faker'
```

## Local filesystem modules

You can import modules on the local file system either 
through relative or absolute paths. Mokapi also supports NodeJS
modules.

```javascript
import { someFunc } from './helpers.js'
import { otherFunc } from '../lib'
import dateTime from 'date-time' // npm install date-time
```

## JSON & YAML file

You can import JSON and YAML file and Mokapi converts the data to a
Javascript object.

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