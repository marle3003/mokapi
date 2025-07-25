---
title: Email Documentation
description: Safely test email functionality without spamming real inboxes using Mokapi’s mock mail server.
---

# Mocking Mail Server

## Introduction

The Mokapi Mail Specification provides a structured and declarative way to define a mock email server that
supports both SMTP (for sending messages) and IMAP (for reading them). It is designed to help developers
and testers simulate email workflows in a fully isolated and reliable environment, without relying on
external infrastructure or risking delivery to real inboxes.

With Mokapi's Mail configuration, you can:

- Set up one or more SMTP and IMAP endpoints
- Define user mailboxes with credentials and custom folders
- Create rules to allow or block specific messages
- Test your application’s email flows in a consistent and repeatable way

``` box=info
Mokapi does not send real emails. Instead, it captures and stores incoming messages so they can be 
accessed via IMAP, the Mokapi Dashboard, or programmatically through Mokapi's own API.
```

## Example Configuration

Below is a simple example of a configuration that sets up both SMTP and IMAP servers:

```yaml
mail: '1.0'
info:
  title: Mokapi's Mail Server
servers:
  smtp:
    host: :25
    protocol: smtp
    description: The SMTP server for sending mails
  imap:
    host: :143
    protocol: imap
    description: The IMAP server for reading mails
```

## Specification

This section describes the full structure of a Mokapi mail configuration, including all top-level 
fields and component objects.

### Format

Mail configurations can be written in either YAML or JSON. All field names are case-sensitive.

### Patching Configuration

Mokapi supports patching of mail configuration files, allowing you to override or extend parts of
the base configuration without duplicating the entire file. This is especially useful when you
need to customize settings for different environments (e.g., local, CI, staging) while keeping
your base config clean and reusable.

Learn more about how patching works and how to structure patch files in the [Patching Configuration Guide](/docs/configuration/patching.md).

### Schema Reference

#### Mail Object

This is the root object of the Mail configuration specification. It defines the overall structure, metadata, and key components
of the mock mail server.

| Field Name | Type             | Description                                                                                     |
|------------|------------------|-------------------------------------------------------------------------------------------------|
| mail       | Version String   | **REQUIRED.** Specifies the Mail Specification version being used. Currently, 1.0 is supported. |
| info       | Info object      | **REQUIRED.** Metadata about your mail server.                                                  |
| servers    | Servers object   | Defines SMTP/IMAP endpoints.                                                                    |
| mailboxes  | Mailboxes object | Defines user mailboxes and folders.                                                             |
| settings   | Settings object  | Mail server settings.                                                                           |
| rules      | Rules object     | Optional rules to allow or reject certain messages.                                             |

#### Mail Version String

The version string defines which version of the Mail Specification your configuration adheres to.
It follows the format major.minor (e.g., 1.0).

```yaml
mail: '1.0'
```

#### Info Object

The info object contains human-readable metadata about the mail server configuration, including its title and optional
description or contact information.

| Field Name  | Type           | Description                                                                                            |
|-------------|----------------|--------------------------------------------------------------------------------------------------------|
| title       | string         | **REQUIRED**. The title of the mail server.                                                            |
| description | string         | A short description of the mail server. **CommonMark syntax** can be used for rich text representation |
| version     | string         | Provides the version of the mocked mail server                                                         |
| contact     | Contact Object | The contact information for the mocked mail server                                                     |

##### Info Object Example

```yaml
title: Mail Server
description: This is a sample mocked mail server
```

#### Contact Object

The contact object allows you to specify support or ownership information for the mocked mail server.

| Field Name | Type   | Description                                                                                       |
|------------|--------|---------------------------------------------------------------------------------------------------|
| name       | string | The identifying name of the contact person/organization.                                          |
| url        | string | The URL pointing to the contact information. This MUST be in the form of an absolute URL.         |
| email      | string | The email address of the contact person/organization. MUST be in the format of an email address.  |

##### Contact Object Example

```yaml
name: Mail Support
url: https://www.example.com/support
email: support@example.com
```

#### Servers Object

The servers object defines the SMTP and/or IMAP endpoints exposed by your mock mail server.

| Field Pattern      | Type           | Description                                    |
|--------------------|----------------|------------------------------------------------|
| `^[A-Za-z0-9_-]+$` | Server Object  | Key name for the server and its configuration  |

##### Servers Object Example

```yaml
smtp:
  host: localhost:25
  protocol: smtp
  description: The SMTP server for sending mails
imap:
  host: localhost:143
  protocol: imap
  description: The IMAP server for reading mails
```

#### Server Object

A Server Object describes an individual SMTP or IMAP server, including connection details and protocol information.

| Field Name  | Type   | Description                                                                      |
|-------------|--------|----------------------------------------------------------------------------------|
| host        | string | **REQUIRED**. The server host name. It MAY include the port.                     |
| protocol    | string | **REQUIRED**. The supported protocol: *smtp*, *imap*, *smtps*, or *imaps*.       |
| description | string | Optional description for documentation purposes. Supports **CommonMark syntax**. |

##### Server Object Example

```yaml
host: localhost:25
protocol: smtp
description: The SMTP server for sending mails
```

#### Mailboxes Object

Defines mock email accounts. Each entry represents a full email address and its corresponding settings.

| Field Pattern  | Type           | Description                                              |
|----------------|----------------|----------------------------------------------------------|
| *mailbox name* | Mailbox Object | The definition of a mailbox, keyed by the email address  |

##### Mailboxes Object Example

```yaml
bob@example.com:
  username: bob
  password: bob123
  folders:
    Trash:
      flags: [\Trash]
alice@example.com:
  username: alice
  password: alice123
```

#### Mailbox Object

A Mailbox Object contains user credentials and a nested folder structure for accessing stored messages.

| Field Name | Type            | Description                                        |
|------------|-----------------|----------------------------------------------------|
| username   | string          | Optional username used to authenticate             |
| password   | string          | Optional password used to authenticate             |
| folders    | Folders Object  | Optional map of folders belonging to this mailbox  |

##### Mailbox Object Example

```yaml
username: bob
password: bob123
folders:
  Trash:
    flags: [\Trash]
    folders:
      2024: {}
```

#### Folders Object

The folders object is a map of folder names to Folder Objects. Folders may also contain nested folders, allowing
deep hierarchical structures.

| Field Pattern | Type          | Description                                                          |
|---------------|---------------|----------------------------------------------------------------------|
| *folder name* | Folder Object | The definition of a mailbox folder, keyed by the name of the folder  |


#### Folder Object

Each Folder Object defines an individual folder’s behavior and structure.

| Field Name | Type            | Description                                       |
|------------|-----------------|---------------------------------------------------|
| flags      | string          | Optional list of IMAP flags (e.g., \Trash, \Sent) |
| folders    | Folders Object  | Optional map of folders nested within this one    |

##### Folder Object Example

```yaml
flags: [\Trash]
folders:
  2024: {}
```

#### Settings Object

The settings object defines global configuration options that influence server behavior.

| Field Name        | Type    | Default | Description                                                                                                                                                                                        |
|-------------------|---------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| maxRecipients     | integer | 0       | Optional maximum number of recipients per email. Use 0 for unlimited.                                                                                                                              |
| autoCreateMailbox | boolean | true    | Optional setting to auto-generate mailboxes at runtime.                                                                                                                                            |
| maxInboxMails     | integer | 100     | (Optional) Maximum number of messages kept in the INBOX folder. Oldest mails are removed when the limit is exceeded. Use 0 to disable the limit and store messages indefinitely (not recommended). |

##### Mailbox Object Example

```yaml
maxRecipients: 20
autoCreateMailbox: true
```

#### Rules Object

The rules object defines message acceptance policies for the mock server.
Each rule is keyed by a unique name and specifies criteria for allowing or denying incoming messages.

| Field Pattern      | Type        | Description              |
|--------------------|-------------|--------------------------|
| `^[A-Za-z0-9_-]+$` | Rule Object | The definition of a rule |

##### Rules Object Example

```yaml
allowSender:
  sender: alice@example.com
  action: allow
denySubject:
  subject: Hello.*
  action: deny
```

#### Rule Object

A Rule Object defines a pattern-based rule that determines whether a message is accepted or rejected.
The value of each matching field (e.g., sender, subject, body) can be either an exact string or a regular
expression pattern for more flexible filtering.

| Field Name      | Type                   | Description                         |
|-----------------|------------------------|-------------------------------------|
| sender          | string                 | Match against the sender's address  |
| recipient       | string                 | Match against recipient's address   |
| subject         | string                 | Match subject line                  |
| body            | string                 | Match against message body          |
| action          | string                 | Required: either allow or deny      |
| rejectResponse  | Reject Response Object | Optional custom rejection response  |

##### Server Object Example

```yaml
sender: alice@example.com
action: allow
```

#### Reject Response Object

A Reject Response Object defines a detailed response when an incoming message is denied based on a rule.

| Field Name         | Type                   | Description                                       |
|--------------------|------------------------|---------------------------------------------------|
| statusCode         | integer                | SMTP status code to return (e.g., 550).           |
| enhancedStatusCode | string                 | Extended SMTP status code (e.g., 5.7.1).          |
| message            | string                 | Human-readable message explaining the rejection.  |

##### Reject Response Object Example

```yaml
statusCode: 550
enhancedStatusCode: 5.7.1
text: Your error message
```