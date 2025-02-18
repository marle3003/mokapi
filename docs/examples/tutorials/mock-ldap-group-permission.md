---
title:  Mocking LDAP with Group Permissions in a Node.js Backend
description: Learn how to mock LDAP authentication and group permission using Mokapi and a Node.js backend.
icon: bi-shield-lock
---

# Mocking LDAP with Group Permissions in a Node.js Backend

## Introduction

In this tutorial, you'll learn how to use Mokapi to mock an LDAP server for authentication and group-based access 
control in a Node.js backend. This is useful for testing without relying on a real LDAP 
server.

### What You'll Learn

- Set up a mock LDAP server using Mokapi
- Authenticate users via LDAP in a Node.js backend
- Implement group-based permissions

## 1. Create an LDAP Mock Configuration

Create a file named `ldap.yaml` with the following content defining on which port the server
should listen and LDIF files to import. 

```yaml tab=ldap.yaml
ldap: 1.0.0
info:
  title: Mokapi's LDAP Server
  description: An example configuration to mock an LDAP server
host: :389
files:
  - ./users.ldif
```

Next create the `users.ldif` file with the users and group:

```ldif tab=users.ldif
dn: dc=mokapi,dc=io

dn: uid=awilliams,dc=mokapi,dc=io
cn: Alice Williams
uid: awilliams
userPassword: foo123

dn: uid=bmiller,dc=mokapi,dc=io
cn: Bob Miller
uid: bmiller
userPassword: bar123

dn: uid=csmith,dc=mokapi,dc=io
cn: Carol Smith
uid: csmith
userPassword: secret123
memberOf: cn=Admins,ou=groups,dc=mokapi,dc=io

dn: ou=groups,dc=mokapi,dc=io

dn: cn=Admins,ou=groups,dc=mokapi,dc=io
cn: "Admins"
```

Here, we define a user Carol Smith, who is member of the Admins group.

Using [LDAP Browser](https://marketplace.visualstudio.com/items?itemName=fengtan.ldap-explorer) in VSCode the directory would look like this:

<img src="./vscode-ldap-browse-example.png" alt="Screenshot of VSCode displaying an LDAP directory structure." />

## 2. Implementing Backend in Node.js

Create a file `auth.js` and add:

```javascript
const ldap = require('ldapjs');

const client = ldap.createClient({
    url: 'ldap://localhost:389'
});

function authenticate(username, password, callback) {
    const dn = `uid=${username},${BASE_DN}`;
    client.bind(dn, password, (err) => {
        if (err) return callback(null, false);
        callback(null, true);
    });
}

function checkGroup(username, groupName, callback) {
    const searchOptions = {
        filter: `(memberOf=cn=${groupName},ou=groups,${BASE_DN})`,
        scope: 'sub'
    };

    client.search(`cn=${username},${BASE_DN}`, searchOptions, (err, res) => {
        let found = false;
        res.on('searchEntry', () => found = true);
        res.on('end', () => callback(null, found));
    });
}

module.exports = { authenticate, checkGroup };
```

In 'server.js' we use the LDAP authentication and group-based access control:

```javascript tab=server.js
const express = require('express');
const { authenticate, checkGroup } = require('./auth');

const PORT = 3000;

const app = express();
app.use(express.json());

app.post('/login', (req, res) => {
    const { username, password } = req.body;
    authenticate(username, password, (err, success) => {
        if (!success) return res.status(401).send('Invalid credentials');
        checkGroup(username, 'Admins', (err, isAdmin) => {
            if (isAdmin) {
                res.send({ message: 'Welcome, Admin!' });
            } else {
                res.status(403).send('Access denied');
            }
        });
    });
});

app.listen(PORT, () => console.log(`Server running on port ${PORT}`));
```

## 3. Testing the Mock LDAP Authentication 

Start Mokapi LDAP server:

```shell
mokapi ldap.yaml
```

Then, start the backend server:

```shell
node server.js
```

Now, test the login with cURL or [Bruno](https://www.usebruno.com/) 

```shell
curl -X POST http://localhost:3000/login -H "Content-Type: application/json" -d '{"username": "csmith", "password": "secret123"}'
```

Expected Response, if user is in Admins group:

```json
{ "message": "Welcome, Admin!" }
```

If the user is not in the Admins group, the response will be:

```json
{ "message": "Access denied" }
```

If the username and/or password is not correct, the response will be:

```json
{ "message": "Invalid credentials" }
```

## Conclusion

You’ve successfully mocked an LDAP server with Mokapi, authenticated users, and 
implemented group-based permissions in a Node.js backend!