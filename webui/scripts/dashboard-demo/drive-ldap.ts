import { Client, Change, Attribute } from 'ldapts';

const client = new Client({ url: 'ldap://localhost:8389' });

export async function driveLdap() {
    await client.bind('dc=hr,dc=example,dc=com');

    await client.search('dc=hr,dc=example,dc=com', {
        filter: '(uid=ajohnson)',
        scope: 'sub'
    });

    await client.search('dc=hr,dc=example,dc=com', {
        filter: '(&(objectCategory=user)(memberOf=cn=Sales,ou=departments,dc=hr,dc=example,dc=com))',
        scope: 'sub'
    });

    await client.modify('uid=bmiller,ou=people,dc=hr,dc=example,dc=com',
        new Change({ 
            operation: 'add', 
            modification: new Attribute({
                type: 'telephoneNumber', values: ['+1 555 123 9876'] 
            })
        }),
    );

    await client.unbind();
}