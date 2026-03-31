import { on } from 'mokapi'
import { base64 } from 'mokapi/encoding'

export default function() {
    on('http', function(request, response) {
        const authHeader = request.header['Authorization'];
        if (authHeader && authHeader.startsWith('Bearer ')) {
            const token = authHeader.substring(7);
            try {
                const { payload } = decodeJWT(token)
                if (payload.scopes && payload.scopes.includes('read')) {
                    console.log('Access granted: read scope');
                } else {
                    response.statusCode = 403;
                    response.data = { error: 'Insufficient scope' };
                }
            } catch (err) {
                response.statusCode = 401;
                response.data = { error: 'Invalid token' };
            }
        } else {
            response.statusCode = 401;
            response.data = { error: 'Unauthorized' };
        }
        return true
    });
};

function decodeJWT(jwt) {
    const parts = jwt.split('.');
    console.log(parts.length)
    if (parts.length !== 3) {
        throw new Error('Invalid JWT format');
    }
    const [encodedHeader, encodedPayload, signature] = parts;
    const header = JSON.parse(base64UrlDecode(encodedHeader));
    const payload = JSON.parse(base64UrlDecode(encodedPayload));
    return { header, payload, signature };
}

// Base64URL decode function
function base64UrlDecode(base64Url) {
    let b = base64Url.replace(/-/g, '+').replace(/_/g, '/');  // Replace URL-safe chars with standard Base64
    const padding = base64.length % 4 === 0 ? '' : '='.repeat(4 - base64.length % 4); // Add padding if needed
    return base64.decode(b + padding)
}