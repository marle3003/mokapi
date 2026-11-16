import { app } from 'mokapi'

export default function () {
    app.http()
        .operation('serviceLdap', (req, res) => {
            res.data = servers.find(x => x.name === req.path['name']) ?? servers[0]
        })
}

export const servers = [
    {
        name: "LDAP Testserver",
        description: "This is a sample LDAP server",
        version: "1.0",
        server: "0.0.0.0:389",
    },
    {
        name: "LDAP Second Server",
        description: "Another sample LDAP server",
        version: "1.0",
        server: "0.0.0.0:636",
    }
]

export const searches = [
    {
        id: "a1289b9b-aff7-4c53-92a0-808c7ce7d907",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver",
            operation: "search"
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
            name: "LDAP Testserver",
            operation: "modify"
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
            name: "LDAP Testserver",
            operation: "add"
        },
        time:  '2025-02-27T11:50:25.482366+01:00',
        data: {
            duration: 1,
            request: {
                operation: 'Add',
                dn: 'cn=alice,cn=users,dc=mokapi,dc=io',
                attributes: [
                    {
                        type: 'sn',
                        values: [ 'Smith' ]
                    },
                    {
                        type: 'objectClass',
                        values: [ 'top', 'person' ]
                    },
                    {
                        type: 'cn',
                        values: [ 'alice' ]
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
            name: "LDAP Testserver",
            operation: "delete"
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
                matchedDn: 'cn=users,dc=mokapi,dc=io',
                message: 'delete operation failed: the specified entry does not exist: cn=turangal,cn=users,dc=mokapi,dc=io'
            }
        }
    },
    {
        id: "2fa8df80-9b15-427b-9c6d-0503aec06ed3",
        traits: {
            namespace: "ldap",
            name: "LDAP Testserver",
            operation: "modifydn"
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
            name: "LDAP Testserver",
            operation: "compare"
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
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000001",
        traits: { namespace: "ldap", name: "LDAP Testserver", operation: "bind" },
        time: '2025-02-27T11:48:50.482366+01:00',
        data: {
            duration: 2,
            request: {
                operation: 'Bind',
                version: '3',
                name: 'cn=admin,dc=mokapi,dc=io',
                password: 'secret',
                auth: 'simple'
            },
            response: {
                status: 'Success',
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000002",
        traits: { namespace: "ldap", name: "LDAP Testserver", operation: "unbind" },
        time: '2025-02-27T11:48:40.482366+01:00',
        data: {
            duration: 0,
            request: {
                operation: 'Unbind',
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000003",
        traits: { namespace: "ldap", name: "LDAP Testserver", operation: "search" },
        time: '2025-02-27T11:48:30.482366+01:00',
        data: {
            duration: 12,
            request: {
                operation: 'Search',
                baseDN: 'dc=mokapi,dc=io',
                scope: 'WholeSubtree',
                sizeLimit: 0,
                timeLimit: 0,
                filter: '(cn=alice)',
                attributes: [ 'cn', 'sn', 'objectClass' ]
            },
            response: {
                status: 'Success',
                results: [
                    {
                        dn: 'cn=alice,cn=users,dc=mokapi,dc=io',
                        attributes: {
                            cn: [ 'alice' ],
                            sn: [ 'Smith' ],
                            objectClass: [ 'top', 'person' ]
                        }
                    },
                    {
                        dn: 'cn=alice2,cn=users,dc=mokapi,dc=io',
                        attributes: {
                            cn: [ 'alice2' ],
                            sn: [ 'Jones' ],
                            objectClass: [ 'top', 'person' ]
                        }
                    }
                ]
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000004",
        traits: { namespace: "ldap", name: "LDAP Testserver", operation: "search" },
        time: '2025-02-27T11:48:20.482366+01:00',
        data: {
            duration: 8,
            request: {
                operation: 'Search',
                baseDN: 'dc=mokapi,dc=io',
                scope: 'BaseObject',
                sizeLimit: 1,
                timeLimit: 10,
                filter: '(objectClass=*)',
                attributes: []
            },
            response: {
                status: 'Success',
                results: [
                    {
                        dn: 'dc=mokapi,dc=io',
                        attributes: {}
                    }
                ]
            },
            actions: [
                {
                    duration: 1500000,
                    tags: {
                        name: "dashboard",
                        file: "examples/mokapi/ldap_handler.js",
                        fileKey: "c7d6b1a2-5f0e-4c4a-9d0a-1b3c8e7f2a10",
                        event: "ldap"
                    },
                    logs: [
                        { level: 'info', message: 'a custom log message from the ldap handler' }
                    ],
                    parameters: [
                        JSON.stringify({
                            operation: 'Search',
                            baseDN: 'dc=mokapi,dc=io',
                            scope: 'BaseObject',
                            filter: '(objectClass=*)'
                        }),
                        JSON.stringify({
                            status: 'Success',
                            results: [ { dn: 'dc=mokapi,dc=io', attributes: {} } ]
                        })
                    ]
                }
            ]
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000005",
        traits: { namespace: "ldap", name: "LDAP Testserver", operation: "add" },
        time: '2025-02-27T11:48:10.482366+01:00',
        data: {
            duration: 1,
            request: {
                operation: 'Add',
                dn: 'cn=bob,cn=users,dc=mokapi,dc=io',
                attributes: [
                    {
                        type: 'cn',
                        values: [ 'bob' ]
                    }
                ]
            },
            response: {
                status: 'EntryAlreadyExists',
                message: 'add operation failed: entry already exists: cn=bob,cn=users,dc=mokapi,dc=io',
                matchedDn: 'cn=users,dc=mokapi,dc=io'
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000006",
        traits: { namespace: "ldap", name: "LDAP Testserver", operation: "search" },
        time: '2025-02-27T11:48:00.482366+01:00',
        data: {
            duration: 3,
            request: {
                operation: 'Search',
                baseDN: 'cn=users,dc=mokapi,dc=io',
                scope: 'SingleLevel',
                sizeLimit: 10,
                timeLimit: 30,
                filter: '(objectClass=group)',
                attributes: []
            },
            response: {
                status: 'Success',
                results: []
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000007",
        traits: { namespace: "ldap", name: "LDAP Second Server", operation: "bind" },
        time: '2025-02-27T11:47:50.482366+01:00',
        data: {
            duration: 2,
            request: {
                operation: 'Bind',
                version: '3',
                name: 'cn=reader,dc=example,dc=com',
                password: 'reader',
                auth: 'simple'
            },
            response: {
                status: 'Success',
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000008",
        traits: { namespace: "ldap", name: "LDAP Second Server", operation: "search" },
        time: '2025-02-27T11:47:40.482366+01:00',
        data: {
            duration: 4,
            request: {
                operation: 'Search',
                baseDN: 'dc=example,dc=com',
                scope: 'WholeSubtree',
                sizeLimit: 10,
                timeLimit: 30,
                filter: '(uid=reader)',
                attributes: []
            },
            response: {
                status: 'Success',
                results: [
                    {
                        dn: 'uid=reader,dc=example,dc=com',
                        attributes: {}
                    }
                ]
            }
        }
    },
    {
        id: "b1c1a001-0000-4000-8000-000000000009",
        traits: { namespace: "ldap", name: "LDAP Second Server", operation: "unbind" },
        time: '2025-02-27T11:47:30.482366+01:00',
        data: {
            duration: 0,
            request: {
                operation: 'Unbind',
            }
        }
    }
]
