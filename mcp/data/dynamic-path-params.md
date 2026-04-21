# Scenario dynamic-path-params

HTTP mock handler to get a pet stored in an array list.
Demonstrates how to:
- Access request parameters
- Apply custom logic (e.g., lookup, filtering)

```typescript
import { on } from "mokapi"

let pets = [
  { id: 1, name: 'Fluffy', status: 'available', category: { id: 1, name: 'Dogs' }, photoUrls: [], tags: [] },
  { id: 3, name: 'Hedgie', status: 'pending', category: { id: 2, name: 'Small Animals' }, photoUrls: [], tags: [] }
];

export default function () {
  on('http', async (request, response) => {
    switch(request.key) {
      case '/pets/{id}':
        if (request.method !== 'GET') {
          return
        }
        const pet = pets.find(x => x.id === request.path.id)
        if (pet) {
          response.data = pet
        } else {
          console.log('pet not found', request)
          response.rebuild(404)
        }
	}
  })
}
```