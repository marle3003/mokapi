# Scenario delay-latency

Simulate server latency by delaying the response. Useful to test scenarios:
- frontend loading states
- timeouts
- high-load

```typescript
import { on } from "mokapi"

let pets = [
  { id: 1, name: 'Fluffy', status: 'available', category: { id: 1, name: 'Dogs' }, photoUrls: [], tags: [] },
  { id: 3, name: 'Hedgie', status: 'pending', category: { id: 2, name: 'Small Animals' }, photoUrls: [], tags: [] }
];

export default function () {
  on('http', async (request, response) => {
    switch(request.key) {
      case '/pets': {
        if (request.method !== 'GET') return

        // simulate network latency (e.g., 2 seconds)
        sleep('2s')

        response.data = pets
      }
    }
  })
}
```