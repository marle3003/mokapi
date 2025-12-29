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
        id: "a1289b9b-aff7-4c53-92a0-808c7ce7d907",
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
        id: "fbe9ade9-2adc-451b-a86b-a900c06c0058",
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
        id: "762b4a9e-7f26-4314-8ffd-7ae1922ab330",
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
        id: "23636b50-e19b-4acc-afc0-f5918a7d2e64",
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
        id: "2fa8df80-9b15-427b-9c6d-0503aec06ed3",
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
        id: "34cbe69d-948c-43e3-aa19-df83c09116c4",
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