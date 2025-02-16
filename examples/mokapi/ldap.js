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
        time:  '2023-02-27T11:49:25.482366+01:00',
        data: {
            request: {
                baseDN: 'dc=mokapi,dc=io',
                scope: 2    ,
                sizeLimit: 10,
                filter: '(objectClass=user)',
                attributes: [
                    'memberOf', 'thumbnailphoto'
                ]
            },
            response: {
                status: 'Success',
                results: [
                    {
                        dn: 'CN=turangal,CN=users,DC=mokapi,DC=io',
                        attributes: {}
                    },
                    {
                        dn: 'CN=farnsworthh,CN=users,DC=mokapi,DC=io',
                        attributes: {}
                    },
                    {
                        dn: 'CN=fryp,CN=users,DC=mokapi,DC=io',
                        attributes: {}
                    },
                    {
                        dn: 'CN=wonga,CN=users,DC=mokapi,DC=io',
                        attributes: {}
                    },
                ]
            }
        },
        duration: 10
    }
]