import { Client, Change, Attribute } from 'ldapts';

const client = new Client({ url: 'ldap://localhost:8389' });

export async function driveLdap() {
    await client.bind('dc=hr,dc=example,dc=com');

    await client.search('dc=hr,dc=example,dc=com', {
        filter: '(uid=ajohnson)',
        scope: 'sub'
    });

    await client.search('dc=hr,dc=example,dc=com', {
        filter: '(memberOf=cn=Sales,ou=departments,dc=hr,dc=example,dc=com)',
        scope: 'sub'
    });

    await client.add('uid=cbrown,ou=people,dc=hr,dc=example,dc=com',
        {
            cn: 'Carol Brown',
            uid: 'cbrown',
            userPassword: 'secret790',
            givenName: 'Carol',
            sn: 'Brown',
            objectClass: [ 'top', 'person', 'organizationalPerson', 'inetOrgPerson' ]
        },
    );

    await client.modify('uid=bmiller,ou=people,dc=hr,dc=example,dc=com',
        new Change({ 
            operation: 'add', 
            modification: new Attribute({
                type: 'telephoneNumber', values: ['+1 555 123 9876'] 
            })
        }),
    );

    await client.compare('uid=bmiller,ou=people,dc=hr,dc=example,dc=com', 'telephoneNumber', '+1 555 123 9876')

    await client.modifyDN('uid=cbrown,ou=people,dc=hr,dc=example,dc=com', 'uid=ctaylor,ou=people,dc=hr,dc=example,dc=com');

    await client.del('uid=ctaylor,ou=people,dc=hr,dc=example,dc=com');

    await client.search('dc=hr,dc=example,dc=com', {
        filter: '(userAccountControl:1.2.840.113556.1.4.803:=512)',
        scope: 'sub'
    });

    await client.unbind();
}