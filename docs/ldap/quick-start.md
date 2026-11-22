---
title: "LDAP Mocking: Query and Bind to a Mocked Directory with Mokapi"
description: Start a mocked LDAP directory from a YAML file and an LDIF file, search it and bind as a user with ldapsearch, add an entry and inspect the requests in the dashboard.
---
# Mocking an LDAP Directory

## Overview

This guide mocks a small company directory with Mokapi. By the end, you'll have:

1. Defined a directory with two users and a group in an LDIF file.
2. Started Mokapi so it serves the directory on port 389.
3. Searched the directory with `ldapsearch`.
4. Bound as a user with a correct and a wrong password.
5. Added an entry and seen that a reload resets it.
6. Inspected the requests in the dashboard.

The commands use the OpenLDAP client tools (`ldapsearch`, `ldapadd`). If Mokapi isn't installed yet,
see the [installation guide](/docs/get-started/installation.md).

## Define the Directory

Create the configuration file `ldap.yaml`:

```yaml tab=ldap.yaml
ldap: 1.0 # version of the file format, not the LDAP protocol version
info:
  title: Company Directory
host: :389
server:
  namingContexts:
    - dc=example,dc=com
files:
  - ./users.ldif
```

`host` is the address Mokapi listens on and has no default. `server.namingContexts` sets the naming context that
the root DSE reports. The file in `files` holds the entries:

```ldif tab=users.ldif
dn: dc=example,dc=com
objectClass: top

dn: cn=alice,dc=example,dc=com
objectClass: inetOrgPerson
cn: alice
sn: Smith
userPassword: secret

dn: cn=bob,dc=example,dc=com
objectClass: inetOrgPerson
cn: bob
sn: Jones
userPassword: hunter2

dn: cn=developers,dc=example,dc=com
objectClass: groupOfNames
cn: developers
member: cn=alice,dc=example,dc=com
```

The relative path is resolved against the location of `ldap.yaml`. See the [overview](/docs/ldap/overview.md) for the
complete configuration, including inline entries and URLs as files.

## Start Mokapi

```bash
mokapi ldap.yaml
```

Or run it in Docker. The image's entrypoint is `mokapi`, so the configuration file goes after the image name. Mount
the folder that contains both files:

```bash
docker run --rm -p 389:389 -p 8080:8080 -v "$(pwd):/ldap:ro" mokapi/mokapi:latest /ldap/ldap.yaml
```

Port 389 is the LDAP server and port 8080 the dashboard. If port 389 is already in use or you aren't allowed to bind it, use a port such as `:1389` in `host`, in the
Docker `-p` option and in the commands below.

## Search the Directory

Find all users and return only their `cn` and `sn`:

```bash
ldapsearch -x -H ldap://localhost:389 -LLL -b "dc=example,dc=com" "(objectClass=inetOrgPerson)" cn sn
```

`-x` selects simple authentication and `-LLL` prints plain LDIF without the comments. Mokapi returns the entries in
the order of the LDIF file:

```
dn: cn=alice,dc=example,dc=com
cn: alice
sn: Smith

dn: cn=bob,dc=example,dc=com
cn: bob
sn: Jones
```

The order of the attributes within an entry can differ between requests.

Mokapi calculates `memberOf` from the `member` attributes of the groups, so alice knows her group although the LDIF
file doesn't set it:

```bash
ldapsearch -x -H ldap://localhost:389 -LLL -b "dc=example,dc=com" "(cn=alice)" memberOf
```

```
dn: cn=alice,dc=example,dc=com
memberOf: cn=developers,dc=example,dc=com
```

The root DSE describes the server. Its `namingContexts` attribute is operational, so a search returns it only if you
request it by name or with `+`:

```bash
ldapsearch -x -H ldap://localhost:389 -LLL -s base -b "" namingContexts
```

```
dn:
namingContexts: dc=example,dc=com
```

## Bind as a User

Pass a DN with `-D` and a password with `-w`. Mokapi compares the password with the `userPassword` of the entry:

```bash
ldapsearch -x -H ldap://localhost:389 -LLL -D "cn=alice,dc=example,dc=com" -w secret -b "dc=example,dc=com" "(cn=alice)" cn
```

```
dn: cn=alice,dc=example,dc=com
cn: alice
```

With a wrong password, `ldapsearch` stops at the bind:

```bash
ldapsearch -x -H ldap://localhost:389 -LLL -D "cn=alice,dc=example,dc=com" -w wrong -b "dc=example,dc=com" "(cn=alice)" cn
# ldap_bind: Invalid credentials (49)
```

An unknown DN fails the same way. An entry without a `userPassword`, such as the group, accepts any password.

Mokapi supports only simple authentication, and a bind isn't required for other operations. Without `-D`, the
client binds anonymously and can still search and write.

## Add an Entry

Create `carol.ldif`:

```ldif tab=carol.ldif
dn: cn=carol,dc=example,dc=com
objectClass: inetOrgPerson
cn: carol
sn: Miller
userPassword: s3cret
```

Add it with `ldapadd`:

```bash
ldapadd -x -H ldap://localhost:389 -f carol.ldif
# adding new entry "cn=carol,dc=example,dc=com"
```

A search for `(cn=carol)` now returns the entry, and you can bind as carol. The change only lives in the running
directory. When you save `ldap.yaml` or `users.ldif`, Mokapi rebuilds the directory from the files and carol is gone.
To keep an entry, put it in the LDIF file.

## Inspect Requests in the Dashboard

1. Open the dashboard, by default at http://localhost:8080.
2. Click **LDAP** in the sidebar and then **Company Directory**.
3. The **Requests** sub tab lists every request with operation, DN, status and duration. `ldapsearch` and `ldapadd` also show up as an anonymous bind and an unbind around the actual operation.
4. Click a request to see its details next to the table. For a search, the **Response** sub tab shows the entries
   that Mokapi returned. For the bind with the wrong password, the status is `InvalidCredentials`.

See [Analyze your mocked LDAP API in Dashboard](/docs/ldap/dashboard.md) for filters, pausing the log and sharing a
request by its address.

## Next Steps

- [Mocking an LDAP Server](/docs/ldap/overview.md): Bind rules, search filters, write operations and the full configuration reference.
- [Mock LDAP Authentication in Node.js](/resources/tutorials/mock-ldap-authentication-in-node): Test the login of a Node.js backend against Mokapi.
- [Mocking LDAP using Group Permission](/resources/tutorials/mock-ldap-group-permission-in-node): Test group-based permissions with a mocked directory.
- [JavaScript API](/docs/javascript-api/overview.md): `on('ldap', ...)` lets a script change the response of a search request. It isn't called for bind or write requests.
