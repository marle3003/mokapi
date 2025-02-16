import { metrics } from "./metrics.js"
import { SearchScope, ResultCode } from "mokapi/ldap"

export let server = [
    {
        name: "LDAP Testserver",
        description: "This is a sample LDAP server",
        version: "1.0",
        server: "0.0.0.0:389",
        metrics: metrics.filter(x => x.name.startsWith("ldap"))
    }
]

export const searches = [
    {
        id: "dkads-23124",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver"
        },
        time:  '2025-02-27T11:52:25.482366+01:00',
        data: {
            duration: 30,
            request: {
                operation: 'Search',
                baseDN: 'dc=mokapi,dc=io',
                scope: 'SingleLevel',
                sizeLimit: 10,
                timeLimit: 30,
                filter: '(objectClass=user)',
                attributes: [
                    'memberOf', 'thumbnailphoto'
                ]
            },
            response: {
                status: 'Success',
                results: [
                    {
                        dn: 'cn=turangal,cn=users,dc=mokapi,dc=io',
                        attributes: {}
                    },
                    {
                        dn: 'cn=farnsworthh,cn=users,dc=mokapi,dc=io',
                        attributes: {}
                    },
                    {
                        dn: 'cn=fryp,cn=users,dc=mokapi,dc=io',
                        attributes: {}
                    },
                    {
                        dn: 'cn=wonga,cn=users,dc=mokapi,dc=io',
                        attributes: {}
                    },
                ]
            }
        }
    },
    {
        id: "dkads-10004",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver"
        },
        time:  '2025-02-27T11:50:50.482366+01:00',
        data: {
            duration: 5,
            request: {
                operation: 'Modify',
                dn: 'cn=turangal,cn=users,dc=mokapi,dc=io',
                items: [
                    {
                        modification: 'Add',
                        attribute: {
                            type: 'foo',
                            values: ['bar']
                        }
                    }
                ]
            },
            response: {
                status: 'Success',
            }
        }
    },
    {
        id: "dkads-222",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver"
        },
        time:  '2025-02-27T11:50:25.482366+01:00',
        data: {
            duration: 1,
            request: {
                operation: 'Add',
                dn: 'cn=alice,cn=users,dc=mokapi,dc=io',
                attributes: [
                    {
                        type: 'foo',
                        values: [ 'bar1', 'bar2' ]
                    }
                ]
            },
            response: {
                status: 'Success',
            }
        }
    },
    {
        id: "dwow-12",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver"
        },
        time:  '2025-02-27T11:50:10.482366+01:00',
        data: {
            duration: 1,
            request: {
                operation: 'Delete',
                dn: 'cn=turangal,cn=users,dc=mokapi,dc=io',
            },
            response: {
                status: 'NoSuchObject',
                message: 'delete operation failed: the specified entry does not exist: cn=turangal,cn=users,dc=mokapi,dc=io'
            }
        }
    },
    {
        id: "abc-12",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver"
        },
        time:  '2025-02-27T11:49:25.482366+01:00',
        data: {
            duration: 1,
            request: {
                operation: 'ModifyDN',
                dn: 'cn=turangal,cn=users,dc=mokapi,dc=io',
                newRdn: 'cn=foo',
                deleteOldDn: false,
                newSuperiorDn: 'cn=bar,dc=mokapi,dc=io'
            },
            response: {
                status: 'Success',
            }
        }
    },
    {
        id: "fpp-12",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver"
        },
        time:  '2025-02-27T11:49:10.482366+01:00',
        data: {
            duration: 1,
            request: {
                operation: 'Compare',
                dn: 'cn=turangal,cn=users,dc=mokapi,dc=io',
                attribute: 'foo',
                value: 'bar'
            },
            response: {
                status: 'CompareTrue',
            }
        }
    }
]