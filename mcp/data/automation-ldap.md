# LDAP

Interfaces for exploring LDAP directory

```typescript
interface Ldap extends ApiSummary {
    address: string

    /**
     * Returns all ldap entries.
     */
    getEntries(): LdapEntry[];
}

interface LdapEntry {
    dn: string
    attributes: Record<string, string[]>
}
```