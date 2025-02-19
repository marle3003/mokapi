import { on } from 'mokapi';

export default function() {
    on('http', function(request, response) {
        switch (request.key) {
            case '/protected/apikey':
                if (request.header['API-Key'] === 'my-secret-api-key') {
                    // because mokapi selects first successfull response
                    // we don't need to do anything here
                } else {
                    response.statusCode = 401
                }
                return true
            case '/protected/bearer':
                const auth = request.header['Authorization']
                if (auth && auth.startsWith('Bearer ')) {
                    const token = auth.split(' ')[1];
                    if (token === "valid-jwt-token") {
                        // Token is valid, proceed
                    } else {
                        response.statusCode = 401
                    }
                }
                return true
        }
    })
}