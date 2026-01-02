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
        switch (request.key) {
            case '/user/login': {
                const username = request.query.username;
                if (users.has(username)) {
                    const user = users.get(username);
                    console.log(`[Script] User ${username} found`)
                    if (user.password === request.query.password) {
                        response.data = 'ok'
                    } else {
                        console.log(`[Script] Password for ${username} is invalid`)
                        response.statusCode = 400;
                    }
                } else {
                    console.log(`[Script] User ${username} not found`)
                    response.statusCode = 400;
                }
                break;
            }
            case '/store/order': {
                orders.set(request.body.id, request.body);
                response.data = request.body;
                break;
            }
            case '/store/order/{orderId}': {
                if (request.method === 'GET') {
                    const order = orders.get(request.path.orderId);
                    if (order) {
                        response.data = order;
                    } else {
                        response.statusCode = 404;
                    }
                } else if (request.method === 'DELETE') {
                    const orderId = request.path.orderId;
                    if (orders.delete(orderId)) {
                        response.statusCode = 204;
                    } else {
                        response.statusCode = 404;
                    }
                }
                break;
            }
            case '/user/logout': {
                // force a script error with undefined variable user
                console.log(`logout current user ${user}`)
                break;
            }
        }
    })
}

const users = new Map([
    ['ajohnson', {
        id: fake({ type: 'integer', format: 'int64' }),
        username: 'ajohnson',
        firstName: 'Alice',
        lastName: 'Johnson',
        email: 'alice.johnson@example.com',
        password: 'anothersecretpassword456',
        phone: fake({ type: 'string', pattern: '(?:(?:\\+|0{0,2})91(\\s*[\\- ]\\s*)?|[0 ]?)?[789]\\d{9}|(\\d[ -]?){10}\\d' }),
        userStatus: 1
    }],
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

const orders = new Map([
    [1, {
        petId: 74959,
        quantity: 1,
        shipDate: '2011-01-08T16:59:15Z',
        status: 'delivered'
    }]
])