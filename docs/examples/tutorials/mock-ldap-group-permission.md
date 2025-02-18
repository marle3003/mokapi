---
title:  Mocking LDAP with Group Permissions in a Node.js Backend
description: Learn how to mock LDAP authentication and group permission using Mokapi and a Node.js backend.
icon: bi-shield-lock
---

# Mocking LDAP with Group Permissions in a Node.js Backend

## Introduction

LDAP authentication is widely used for managing user access in enterprise applications. However, setting up a real LDAP 
server for testing can be complex and time-consuming. In this tutorial, you will learn how to use Mokapi to mock an 
LDAP server and implement group-based authentication in a Node.js backend.

### What You'll Learn

- ✅ Setting up a mock LDAP server with Mokapi
- ✅ Authenticating users via LDAP in a Node.js backend
- ✅ Implementing group-based permissions
- ✅ Testing authentication using cURL

## Prerequisites

Before starting, ensure you have the following:

- Node.js installed
- Mokapi installed [Installation Guide](/docs/guides/get-started/installation.md)
- Basic knowledge of LDAP authentication

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

## 2. Implement LDAP Authentication in Node.js

Run the following command in install required dependencies:

```shell
npm install ldapjs express
```

Create a file `auth.js` and add the following code:

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

- **authenticate**: This function binds the provided username and password to the mock LDAP server to verify authentication.
- **checkGroup**: This function searches for the user to determine if they are a member of the Admins group.

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

This Express.js server handles authentication and ensures only users in the Admins group can access protected resources.

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

This setup allows you to test LDAP authentication and group permissions without needing a real LDAP server.