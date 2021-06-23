# LDAP
Use Mokapi to mock an LDAP server

## Example
Following example mocks a Microsoft Active Directory Server

```yaml
ldap: 1.0.0
server:
address: "0.0.0.0:389"
rootDomainNamingContext: DC=mokapi,DC=ch
subSchemaSubentry: CN=schema,DC=mokapi,DC=io
namingContexts:
- DC=mokapi,DC=io
- CN=schema,DC=mokapi,DC=io
entries:
- dn: DC=mokapi,DC=io
  objectClass: top
- dn: CN=users,DC=mokapi,DC=io
  objectClass: top
- dn: CN=farnsworthh,CN=users,DC=mokapi,DC=io
  objectClass: user
  memberOf: CN=Z-Mokapi-Group,DC=mokapi,DC=io
  thumbnailphoto:
  file: ./farnsworth.png
- dn: CN=turangal,CN=users,DC=mokapi,DC=io
  objectClass: user
  memberOf: CN=Z-Mokapi-Group,DC=mokapi,DC=io
  thumbnailphoto:
  file: ./leela.png
- dn: CN=fryp,CN=users,DC=mokapi,DC=io
  objectClass: user
  memberOf: CN=Z-Mokapi-Group,DC=mokapi,DC=io
  thumbnailphoto:
  file: ./philip.png
- dn: CN=wonga,CN=users,DC=mokapi,DC=io
  objectClass: user
  memberOf: CN=Z-Mokapi-Group,DC=mokapi,DC=io
  thumbnailphoto:
  file: ./amy.png
- dn: CN=Z-Mokapi-Group,DC=mokapi,DC=io
  objectClass: group
  objectCategory: group
- dn: CN=schema,DC=mokapi,DC=io
  objectClass:
    - subSchema
    - subEntry
    - top
      objectClasses:
    - ( 2.5.6.0 NAME ( 'top' ) ABSTRACT MUST ( objectClass ) )
    - ( 1.2.840.113556.1.5.9 NAME ( 'user' ) SUP ( top ) MUST ( CN ) MAY ( memberOf ) )
    - ( 1.2.840.113556.1.5.8 NAME ( 'group' ) SUP ( top ) MUST ( CN ) )
      attributeTypes:
    - ( 2.5.21.6 NAME ( 'objectClass' ) DESC 'LDAP object classes' SYNTAX 1.3.6.1.4.1.1466.115.121.1.37 USAGE directoryOperation )
    - ( 1.2.840.113556.1.4.782 NAME ( 'objectCategory' ) SYNTAX 1.3.6.1.4.1.1466.115.121.1.12 USAGE userApplications SINGLE-VALUE )
    - ( 2.5.21.5 NAME ( 'attributeTypes' ) DESC 'LDAP attribute types' SYNTAX 1.3.6.1.4.1.1466.115.121.1.3 USAGE directoryOperation )
    - ( 2.5.4.3 NAME ( 'CN' ) SYNTAX 1.3.6.1.4.1.1466.115.121.1.15 USAGE userApplications )
    - ( 1.2.840.113556.1.2.102 NAME ( 'memberOf' ) SYNTAX 1.3.6.1.4.1.1466.115.121.1.12 USAGE userApplications )
    - ( 2.16.840.1.113730.3.1.35 NAME ( 'thumbnailphoto' ) SYNTAX 1.3.6.1.4.1.1466.115.121.1.40 USAGE userApplications SINGLE-VALUE )
```