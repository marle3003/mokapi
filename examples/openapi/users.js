import {on} from 'mokapi'
import {fake} from 'faker'

export default function() {
    on('http', function(request, response) {
        if (request.url.path === '/api/users') {
            response.data = [
                fake({type: 'string', format: '{username}'}),
                fake({type: 'string', format: '{username}'}),
                fake({type: 'string', format: '{username}'})
            ]
        }
    })
}