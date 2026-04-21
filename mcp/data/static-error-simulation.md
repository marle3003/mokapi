# Scenario static-error-simulation

Return predefined error responses (e.g., 400, 404, 500) for specific endpoints or conditions without dynamic logic.

```typescript
import { on } from "mokapi"

export default function () {
	on('http', (request, response) => {
		switch(request.key) {
			case '/bookings': {
				if (request.method === 'POST') {
					if (request.header['Api-Key'] === 'invalid') {
						// console output will be displayed in the Mokapi's' dashboard
						console.log('api-key is not valid')
						response.rebuild(401)
						return
					}
					if (request.body?.hotel?.code === 'NOT_FOUND') {
						console.log('hotel not found')
						response.rebuild(404)
						return
					}
					if (request.body.hotel.name === 'INVALID') {
						console.log('hotel name is not valid')
						response.rebuild(400)
						return
					}
				}
			}
		}
	})
}
```