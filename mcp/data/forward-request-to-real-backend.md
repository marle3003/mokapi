# Scenario forward-request-to-real-backend

Stop API drift in its tracks. Use Mokapi as a validation layer to enforce OpenAPI contracts between
clients and backends, regardless of who's calling or what they're building. This scenario forwards
incoming requests to real backend services while validating both requests and responses against the
OpenAPI specification.

```typescript
import { on } from 'mokapi';
import { fetch } from 'mokapi/http';

export default async function () {
    
    on('http', async (request, response) => {

        // Map request to backend URL based on OpenAPI spec name
        const url = getForwardUrl(request)

        // If no URL could be determined, return an error immediately
        if (!url) {
            response.statusCode = 500;
            response.body = 'Failed to forward request: unknown backend';
            return;
        } 
            
        try {
            // Forward the request to the backend
            const res = await fetch(url, {
                method: request.method,
                body: request.body,
                headers: request.header,
                timeout: '30s'
            });

            // Copy status code and headers
            response.statusCode = res.statusCode;
            response.headers = res.headers

            // Check the content type to decide whether to validate the response
            const contentType = res.headers['Content-Type']?.[0] || '';

            if (contentType.includes('application/json')) {
                // Mokapi can validate JSON responses automatically
                response.data = res.json();
            } else {
                // For other content types, skip validation
                response.body = res.body;
            }
            
        } catch (e) {
            // Handle any errors that occur while forwarding
            response.statusCode = 500;
            response.body = e.toString();
        }
    });

    function getForwardUrl(request: HttpRequest): string | undefined {
        switch (request.api) {
            case 'backend-1': {
                return 'https://backend1.example.com' + request.url.path + '?' + request.url.query;
			}
			case 'backend-2': {
				return 'https://backend2.example.com' + request.url.path + '?' + request.url.query;
			}
			default:
				return undefined;
		}
	}
}
```