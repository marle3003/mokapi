# Scenario conditional-response

HTTP mock handler for terminals.
Demonstrates how to:
- Access request parameters
- Apply custom logic (e.g., lookup, filtering, updates)

IMPORTANT  
Strict Specification Enforcement:
Mokapi will throw an error if you use a status code NOT defined in the specification.
Always verify the available status codes for each operation before calling response.rebuild() or setting response.statusCode.

```typescript
import { on } from "mokapi"

interface Terminal {
	id: string
	compartments: {
		id: string
		doorState: 'open' | 'closed'
	}[]
}

const terminals: Terminal[] = []

export default function () {
	on('http', (request, response) => {
		switch(request.key) {
			case '/terminals/{id}': {
				const terminal = terminals.find(x => x.id === request.path.id)
				if (!terminal) {
                    // CHECK SPEC: Is 404 defined for this path and HTTP method?
					response.rebuild(404)
					response.data = { error: 'terminal not found' }
					return
				}

				if (request.method === 'GET') {
					response.data = terminal
				} else if (request.method === 'POST') {
					// update the terminal
					Object.assign(terminal, request.body)
					// mokapi already set the success response, nothing to do
				}
				// do not raise an error if different method is used,
				// maybe there is another event handler in a different file defined
				return
			}
			case '/terminals': {
				if (request.method === 'GET') {
					response.data = terminals
				} else if (request.method === 'POST') {
					const terminal = terminals.find(x => x.id === request.path.id)
					if (terminal) {
						// console output will be displayed in the Mokapi's' dashboard
						console.log('terminal already exists', request.body)
                        // CHECK SPEC: Is 400 defined for this path and HTTP method?
						response.rebuild(400)
					} else {
						terminals.push(request.body)
					}
				}
				return
			}
		}
	})
}
```