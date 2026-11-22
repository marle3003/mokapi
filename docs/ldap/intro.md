---
title: Mocking an LDAP Server
description: Mock an LDAP directory with Mokapi to test bind, search and write operations, load entries from LDIF files and reset the directory on reload.
cards:
  items:
    - title: Run your first mocked LDAP
      href: /docs/ldap/quick-start
      description: Start an LDAP mock and query it with ldapsearch.
    - title: Mock LDAP Authentication in Node.js
      href: /resources/tutorials/mock-ldap-authentication-in-node
      description: Test the LDAP login of a Node.js backend against Mokapi.
    - title: Mocking LDAP using Group Permission
      href: /resources/tutorials/mock-ldap-group-permission-in-node
      description: Test group-based permissions of a Node.js backend with a mocked directory.
---

# Mocking an LDAP Server

Mokapi runs an LDAP v3 server that serves a directory you define in a YAML or JSON file and in LDIF files. Your
application connects to it like to any other LDAP server, so you can test login, user lookups and group
membership without a real directory. Clients can also change the directory with add, modify, delete and
modifyDN requests.

## Example Configuration

The configuration file sets the address and lists the LDIF files that hold the entries:

```yaml tab=ldap.yaml
ldap: 1.0 # version of the file format, not the LDAP protocol version
info:
  title: Mokapi's LDAP Server
host: :389
server:
  namingContexts:
    - dc=example,dc=com
files:
  - ./users.ldif
```

```ldif tab=users.ldif
dn: dc=example,dc=com
objectClass: top

dn: cn=alice,dc=example,dc=com
objectClass: inetOrgPerson
cn: alice
sn: Smith
userPassword: secret
```

Start Mokapi with `mokapi ldap.yaml` and query it:

```bash
ldapsearch -x -H ldap://localhost:389 -b "dc=example,dc=com" -D "cn=alice,dc=example,dc=com" -w secret "(cn=alice)"
```

The [Quick Start](/docs/ldap/quick-start.md) walks through this setup step by step. Mokapi watches the
configuration and LDIF files and applies changes without a restart. Relative file paths are resolved against the
configuration file, and a file can also be an HTTP or Git URL.

## How the Mock Behaves

### Root DSE and schema

If you don't define a root DSE (an entry with an empty DN), Mokapi creates one with `supportedLDAPVersion`,
`vendorName`, `vendorVersion`, `namingContexts` (`dc=mokapi,dc=io` by default) and `subschemaSubentry`. If your LDIF
file defines the root DSE, Mokapi keeps your values and only fills in the attributes that are missing. Set
`server.namingContexts` to override the naming contexts in either case.

The entry named by `subschemaSubentry` (default `cn=subschema`) is the schema. If it contains `attributeTypes` and
`objectClasses`, Mokapi validates entries that clients add or modify against them (see Write operations). The schema
entry is only returned by a search when it is the search base.

### Bind

Mokapi supports LDAP version 3 and simple authentication only. Other versions fail with `protocolError (2)` and other
authentication methods, such as SASL, with `authMethodNotSupported (7)`.

A bind with an empty DN is an anonymous bind and succeeds. For a DN, Mokapi looks up the entry and compares the
password with the first `userPassword` value as plain text. An unknown DN or a wrong password fails with
`invalidCredentials (49)`. An entry without a `userPassword` attribute accepts any password.

A bind isn't required before other operations. Mokapi doesn't check permissions, every client can search and write.

### Search

A search supports the scopes base object, single level and whole subtree. Filters can combine `&`, `|` and `!` and use
equality, substring (`cn=al*`), presence (`mail=*`), `>=`, `<=`, approximate (`~=`) and the extensible matching rules
for bitwise AND (`1.2.840.113556.1.4.803`) and bitwise OR (`1.2.840.113556.1.4.804`). Attribute names are case
insensitive. Equality is case insensitive unless the schema defines the attribute with `caseExactMatch`. A filter
Mokapi can't parse fails with `operationsError (1)`.

Entries come back in the order they were loaded. Operational attributes (`namingContexts`, `subschemaSubentry`,
`objectClasses`) are only returned when you request them by name or with `+`. The `memberOf` attribute of a user is
calculated from the `member` attributes of the groups, so it's up to date after every write.

If the request has a size limit, Mokapi stops after that many entries and still reports `success`. The paged results
control (RFC 2696) is supported, the page size and the cookie decide where the next page starts. Other controls are
ignored.

### Write operations

Add, modify, delete and modifyDN change the running directory. Modify supports `add`, `delete` and `replace`.
Failures return these result codes:

| Operation              | Condition                                         | Result code                |
|------------------------|---------------------------------------------------|----------------------------|
| Add                    | The DN already exists                             | `entryAlreadyExists (68)`  |
| Modify, delete, modifyDN | The DN doesn't exist                            | `noSuchObject (32)`        |
| Add, modify            | A value doesn't match the attribute syntax of the schema, or a required attribute of an object class is missing | `constraintViolation (19)` |

Mokapi doesn't check that the parent entry exists when you add an entry, and a delete also works on entries that
have children.

Compare returns `compareTrue (6)` if the entry has the attribute with the value and `compareFalse (5)` in every other
case, also when the entry doesn't exist.

### Reload

A reload of the configuration or of an LDIF file rebuilds the directory from the files. Changes that clients made
with write operations are lost. If several files share the same `info.title`, Mokapi merges them first and the
directory is reset to the merged content.

### Scripts and the dashboard

Use [`on('ldap', ...)`](/docs/javascript-api/mokapi/on.md) to handle search requests in a script. The handler gets
the search request and can replace `response.results`, `response.status` and `response.message`. The type
definitions are `LdapEventHandler`, `LdapSearchRequest` and `LdapSearchResponse` in `@types/mokapi`.

The [Dashboard](/docs/ldap/dashboard.md) lists every request with operation, DN, status and duration.

## Patching

Mokapi patches LDAP configurations by `info.title`. A patch file with the same title as the base file can change
`info.description`, `info.version` and `host`, and adds entries. If an entry already exists, the patch appends its
attribute values to the existing ones instead of replacing them. See
[Patching](/docs/configuration/patching.md) for how Mokapi picks and orders patch files.

## Configuration Reference

Field names are case-sensitive. Mokapi reads YAML and JSON.

### LDAP Object

| Field   | Type          | Description                                                                                                     |
|---------|---------------|-----------------------------------------------------------------------------------------------------------------|
| ldap    | string        | **REQUIRED.** Version of the configuration format. Use `1.0`.                                                  |
| info    | Info object   | **REQUIRED.** Metadata of the directory. `info.title` identifies the directory.                                 |
| host    | string        | Address Mokapi listens on, for example `:389` or `127.0.0.1:1389`. Takes precedence over `server.address`.     |
| server  | Server object | Settings of the root DSE. `server.address` is an alternative to `host`.                                         |
| files   | string array  | Paths or URLs of LDIF files. Mokapi applies them in the given order.                                            |
| entries | array         | Entries defined inline. Each item has a `dn` and the attributes as strings or lists of strings.                 |

The configuration has no default for `host`, so set it explicitly.

### Info Object

| Field       | Type   | Description                                                                         |
|-------------|--------|-------------------------------------------------------------------------------------|
| title       | string | **REQUIRED.** Name of the directory. Also used to merge and patch files, and shown as `dsServiceName` in the root DSE. |
| description | string | Description of the directory. Also set as `description` in the root DSE.            |
| version     | string | Version of the mocked directory.                                                    |

### Server Object

| Field                   | Type         | Description                                                                          |
|-------------------------|--------------|--------------------------------------------------------------------------------------|
| address                 | string       | Address Mokapi listens on, if `host` isn't set.                                      |
| namingContexts          | string array | Value of `namingContexts` in the root DSE. Default: `dc=mokapi,dc=io`.               |
| rootDomainNamingContext | string       | Value of `rootDomainNamingContext` in the root DSE.                                  |

### Inline Entries

Instead of LDIF you can define entries in the configuration file:

```yaml
ldap: 1.0
info:
  title: Inline LDAP Server
host: :389
entries:
  - dn: cn=alice,dc=example,dc=com
    objectClass: [inetOrgPerson]
    sn: Smith
    userPassword: secret
```

Attribute values must be strings, so quote numbers such as `uidNumber: "1000"`. Other value types are ignored.

Inline entries are loaded first. If an LDIF record adds a DN that already exists, Mokapi can't load the configuration
and reports `already exists`.

### LDIF Files

Mokapi reads LDIF records without a `changetype` (or with `changetype: add`) as new entries and also supports
`changetype: modify`, `delete` and `modrdn`, so a later file can change entries of an earlier one. Attribute values
can be read from a file with `attribute:< file:///path` or given as base64 with `attribute:: value`. An empty
`dn:` line starts the root DSE.

{{ card-grid key="cards" }}
