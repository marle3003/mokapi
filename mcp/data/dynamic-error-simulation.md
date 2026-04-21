# Scenario dynamic-error-simulation

Return error responses based on runtime conditions, such as missing resources, validation failures, or conflicting state.

```typescript
import { on } from "mokapi"

const hotels = []

export default function () {
	on('http', (request, response) => {
		switch(request.key) {
			case '/bookings': {
				const hotel = hotels.find(x => x.code === request.body?.hotel?.code)

				if (!hotel) {
				  console.log('hotel not found')
				  response.rebuild(404)
				  response.data = { error: 'hotel not found' }
				  return
				}

				// simulate dynamic errors based on hotel simulation config
				const type = hotel.simulation?.responseType
				switch (type) {
					case 'bad-request':
						response.rebuild(400)
						return
					case 'unauthorized':
						response.rebuild(401)
						return
					case 'forbidden':
						response.rebuild(403)
						return
					case 'internal-server-error':
						response.rebuild(500)
						return
				}
				// success path: generate valid response
				// ...
			}
		}
	})
}
```