---
title: Mock LDAP Authentication in Node.js
description: Learn how to mock LDAP authentication using Mokapi and a Node.js backend. Step-by-step guide with code examples for testing LDAP login without a real server!
icon: bi-person-check
tech: ldap
---

# Mock LDAP Authentication in Node.js

In this tutorial, we'll guide you through the process of setting up a mock LDAP server using Mokapi and creating a 
Node.js backend that authenticates users against it. This approach allows you to test LDAP authentication without 
the need for a real LDAP server. Whether you're working on a project that needs LDAP authentication or just want 
to simulate it for testing, this guide will help you get started.

## Step 1: Set Up a Mock LDAP Server with Mokapi

First, create an LDAP configuration file for Mokapi, e.g., ldap.yaml. This file will define the mock LDAP server's 
settings and the users it will recognize.

```yaml tab=ldap.yaml
ldap: 1.0.0
info:
  title: Mokapi's LDAP Server
  description: An example configuration to mock an LDAP server
host: :389
files:
  - ./users.ldif
```

- **info**: Describes the mock LDAP server.
- **host**: Defines the server’s address and port (default LDAP port is 389).
- **files**: Refers to the LDIF file (users.ldif) that contains user entries.

### Create the LDIF File for User Data

Next, create an LDIF file which contains user entries. This file will define the mock users and their passwords.

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
```

### Explanation

- **dn** (Distinguished Name): Defines the unique location of each entry in the LDAP tree. Users are stored under dc=mokapi,dc=io.
- **cn** (Common Name): Represents the user's full name.
- **uid** (User ID): A unique identifier for the user.
- **userPassword**: The password associated with the user. Note that, in real-world scenarios, this would typically be hashed.

### Default Root DSE and Naming Contexts

By default, Mokapi’s Root DSE includes the namingContexts attribute, which defines the base DN for the directory. The default namingContexts is:

```ldif
namingContexts: dc=mokapi,dc=io
```

If you want to use a different baseDN (for example, dc=example,dc=com), you must update the Root DSE. You can do this 
by providing an updated ldif file with the correct namingContexts.

### Customizing the Root DSE

To customize the Root DSE for a different base DN, you can modify the users.ldif file like this:

```ldif
dn:
namingContexts: dc=example,dc=com
```

### Run Mokapi

To start the mock LDAP server, run the following command:

```bash
mokapi ldap.yaml 
```

This will launch Mokapi with the provided configuration, simulating an LDAP server on localhost:389.

## Step 2: Create a Node.js Backend for LDAP Authentication

Now, let’s build a simple Node.js backend that will authenticate users against the mock LDAP server. This backend 
will accept login requests and validate credentials using the mock LDAP server.

### Install Dependencies

Install the necessary dependencies:

```bash
npm install express ldapjs
```

- **express**: A minimal web framework for Node.js.
- **ldapjs**: A library that allows us to interact with LDAP servers.

## Implement LDAP Authentication in Node.js

Create a server.js file in your project folder with the following code:

```javascript tab=server.js
const express = require("express");
const ldap = require("ldapjs");

const app = express();
app.use(express.json());

const LDAP_URL = 'ldap://localhost:389';
const BASE_DN = 'dc=mokapi,dc=io';

function authenticate(username, password, callback) {
    const client = ldap.createClient({ url: LDAP_URL });
    const userDn = `uid=${username},${BASE_DN}`;

    client.bind(userDn, password, (err) => {
        client.unbind();
        if (err) {
            return callback(false);
        }
        callback(true);
    });
}

app.post("/login", (req, res) => {
    const { username, password } = req.body;
    authenticate(username, password, (success) => {
        if (success) {
            res.json({ message: "Authentication successful" });
        } else {
            res.status(401).json({ message: "Invalid credentials" });
        }
    });
});

const PORT = 3000;
app.listen(PORT, () => {
    console.log(`Server running on http://localhost:${PORT}`);
});
```

#### Code Explanation:
- **authenticate()**: This function binds to the mock LDAP server using the provided username and password. If the bind operation is successful, it returns true, otherwise false.
- **/login endpoint**: A POST route that accepts a username and password in the request body, calls authenticate(), and responds accordingly.
- **Express server**: The app listens on port 3000 for incoming requests.

## Step 3: Test LDAP Authentication

Run the Node.js server by executing:

```shell
node server.js
```

### Send a Login Request

To test the authentication, use a tool like cURL or [Bruno](https://www.usebruno.com/) to send a POST request:

```shell
curl -X POST http://localhost:3000/login \
     -H "Content-Type: application/json" \
     -d '{"username": "awilliams", "password": "foo123"}'
```

### Expected response:

Successful authentication (valid credentials):

```json
{
  "message": "Authentication successful"
}
```

Failed authentication (incorrect password or user not found):

```json
{
  "message": "Invalid credentials"
}
```

## Conclusion
You've successfully set up a mock LDAP server using Mokapi and created a Node.js backend for testing LDAP 
authentication. This mock LDAP server can now be used in your application or tests, simulating real LDAP 
authentication without requiring a full LDAP server.