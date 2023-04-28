import { on } from 'mokapi'

export default function() {
    on('ldap', function(request, response){
        if (request.filter === '(objectClass=foo)'){
            response.results = [
                {
                    dn: 'CN=bob,CN=users,DC=mokapi,DC=io',
                    attributes: {
                        mail: ['bob@mokapi.io'],
                        objectClass: ['foo']
                    }
                }
            ]
            return true
        }
    })
}