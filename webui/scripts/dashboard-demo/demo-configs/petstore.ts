import { on } from 'mokapi';
import { fake } from 'mokapi/faker'

export default function() {
    on('http', (request, response) => {
        if (request.operationId === 'getUserByName') {

            const username = request.path.username;
            console.log(`[HTTP] Incoming request → GET /user/${username}`);
            console.log(`[HTTP] Operation matched: getUserByName`);
            console.log(`[Script] Looking up user: "${username}"`);

            if (users.has(username)) {
                console.log(`[Script] ✓ User found`);
                response.data = users.get(request.path.username);
            } else {
                console.log(`[Script] ✗ User not found → returning 404`);
                response.statusCode = 404
            }
        }
    })
}

const users = new Map([
    ['bmiller', {
        id: fake({ type: 'integer', format: 'int64' }),
        username: 'bmiller',
        firstName: 'Bob',
        lastName: 'Miller',
        email: 'bob.miller@example.com',
        password: 'mysecretpassword123',
        phone: fake({ type: 'string', pattern: '(?:(?:\\+|0{0,2})91(\\s*[\\- ]\\s*)?|[0 ]?)?[789]\\d{9}|(\\d[ -]?){10}\\d' }),
        userStatus: 1
    }]
]);