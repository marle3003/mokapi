---
title: LDAP Documentation
description: Integrate your App with a mock LDAP server
---

# Mocking LDAP server

Mokapi simplifies the process of creating a basic read-only LDAP server mock, enabling you to efficiently test and debug your LDAP interactions. Whether your goal is to validate query accuracy or simulate specific LDAP entry scenarios, Mokapi provides a flexible and developer-friendly solution.

Designed with integration in mind, Mokapi makes LDAP directories accessible and seamlessly integrates into your existing codebases and projects. It empowers developers to experiment, prototype, and troubleshoot without relying on a live LDAP environment.

With custom Mokapi scripts, you gain full control over the responses sent to search requests. This flexibility allows you to tailor the behavior of your mock server to suit your testing needs. Additionally, Mokapi supports the import of LDAP entries through LDIF files, streamlining the setup process and enabling you to test with real or representative data effortlessly.

By leveraging Mokapi, you can enhance your development workflow, reduce dependencies on external systems, and create a smoother path to delivering robust LDAP integrations.

Learn how to create your first LDAP mock with Mokapi and begin ensuring the reliability and robustness of your application.

## Before you start

There are various ways to run Mokapi depending on your needs. For detailed instructions on how to get Mokapi running on your workstation, refer to the information provided [here](/docs/guides/get-started/running.md).

## Basic structure of an LDAP server configuration

To run an LDAP server with Mokapi, the minimum requirement is specifying a host address. This allows Mokapi to bind to the correct network interface and start serving requests.

Below is an example of the simplest configuration, which does not include any custom LDAP entries. In this configuration, Mokapi automatically generates a basic Root DSE entry, which provides essential information about the server, such as the supported LDAP versions, the vendor name and vendor version.

This basic setup is ideal for testing and development purposes, where the focus is on simulating server behavior rather than interacting with actual LDAP entries. You can later extend this configuration by adding LDAP entries and more advanced server settings as needed.

```yaml tab=ldap.yaml
ldap: 1.0 # file configuration version not LDAP protocol version
host: :389
```

To start Mokapi with a specific configuration file, you can use the --provider-file-filename option in the command line. This tells Mokapi to load the specified configuration file when it starts.

```bash
mokapi --providers-file-filename ldap.yaml
```

## Setup a Simple LDAP Entry Structure

Mokapi allows you to configure your LDAP mock server using LDIF files. It supports a wide range of LDIF operations,
such as adding new entries, modifying attributes, and even deleting attributes. This flexibility makes it
easy to simulate real-world LDAP scenarios and test different interactions with your server.

In your configuration file, you can reference multiple LDIF files. Mokapi will continuously monitor these files and
automatically update the LDAP server whenever a change is detected. This dynamic reloading of LDIF files helps streamline
the testing and development process, ensuring that your mock server always reflects the latest configurations.

In the following example, we define an LDAP entry `dc=example,dc=com` and assign it the `top` object class.
This entry is added to the `namingContexts` in the Root DSE. Additionally, we include a user entry for `cn=alice,dc=example,dc=com?`.
This LDIF file can be referenced in the LDAP configuration file using a relative path.

```box=tip
You can also reference an LDIF file using an HTTP or GIT URL, allowing you to source configuration data from remote
locations or version-controlled repositories for better integration and versioning.
```

```ldif tab=example.ldif
# Root DSE
dn:
namingContext: dc=example,dc=com

dn: dc=example,dc=com
objectClass: top

dn: cn=alice,dc=example,dc=com
objectClass: inetOrgPerson
cn: alice
```

In the LDAP configuration, you would reference this LDIF file as follows:

```yaml tab=ldap.yaml
ldap: 1.0 # file configuration version not LDAP protocol version
host: :389
files:
  - ./example.ldif
```

This setup provides a foundational LDAP structure, which you can build upon by adding more entries and customizing attributes to meet your testing requirements.

To query the example LDAP setup, you can use the ldapsearch command, which is a common tool for querying LDAP directories from the command line.

```bash
ldapsearch -x -H ldap://localhost:389 -b "dc=example,dc=com" "(objectClass=*)"
# extended LDIF
#
# LDAPv3
# base <dc=example,dc=com> with scope subtree
# filter: (objectClass=*)
# requesting: ALL
#

#
dn:
namingContext: dc=example,dc=com
supportedLDAPVersion: 3
vendorName: Mokapi
vendorVersion: 0.11.4

# example.com
dn: dc=example,dc=com
objectClass: top

# alice, example.com
dn: cn=alice,dc=example,dc=com
objectClass: inetOrgPerson
cn: alice

# search result
search: 2
result: 0 Success
text: Success

# numResponses: 4
# numEntries: 3
```

To query the cn=alice,dc=example,dc=com user entry, you can use a command like the following:

```bash
ldapsearch -x -h localhost -p 389 -b "dc=example,dc=com" "(cn=alice)"
# extended LDIF
#
# LDAPv3
# base <dc=example,dc=com> with scope subtree
# filter: (cn=alice)
# requesting: ALL
#

# alice, example.com
dn: cn=alice,dc=example,dc=com
objectClass: inetOrgPerson
cn: alice

# search result
search: 2
result: 0 Success
text: Success

# numResponses: 2
# numEntries: 1
```
