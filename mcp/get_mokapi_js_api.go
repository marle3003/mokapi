package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerGetMokapiJsAPI(server *mcp.Server) {
	registerTool(server, &mcp.Tool{
		Name: "get_mokapi_js_api",
		Description: `Returns the Mokapi JavaScript API used to implement mock behavior.

Use this to understand:
- HTTP event handling via on('http', handler)
- Kafka event handling via on('kafka', handler)
- Produce Kafka messages to simulate a Kafka event
- Scheduled jobs via every() or cron()
- Utility functions like sleep() and fetch()

When mocking HTTP endpoints:
1. Identify the API, path, method, and response status
2. Call the "generate_http_response" tool
3. Return the generated object directly
Do NOT:
- Manually construct response.data
- Guess response structure
`,
	}, s.GetMokapiJsAPI)
}

func (s *Service) GetMokapiJsAPI(_ context.Context, _ any) (map[string]any, error) {
	return map[string]any{
		"api": map[string]any{

			"events": []any{
				map[string]any{
					"name":        "on",
					"type":        "http",
					"package":     "mokapi",
					"signature":   "on('http', (request, response) => void, args?)",
					"description": "Handles incoming HTTP requests",
					"parameters": []any{
						map[string]any{
							"name":        "request",
							"type":        "object",
							"description": "Contains data of a HTTP request",
							"properties": map[string]any{
								"method": map[string]any{
									"type":        "string",
									"description": "HTTP request method, such as GET or POST",
								},
								"url": map[string]any{
									"type":        "object",
									"description": "Parsed URL of the request",
									"properties": map[string]any{
										"scheme": map[string]any{
											"type": "string",
										},
										"host": map[string]any{
											"type": "string",
										},
										"port": map[string]any{
											"type": "integer",
										},
										"path": map[string]any{
											"type": "string",
										},
										"query": map[string]any{
											"type": "string",
										},
									},
								},
								"key": map[string]any{
									"type":        "string",
									"description": "Path value that matched the OpenAPI path template",
								},
								"operationId": map[string]any{
									"type":        "string",
									"description": "Operation Id defined in the OpenAPI specification",
								},
								"path": map[string]any{
									"type":        "object",
									"description": "Path parameters defined by the OpenAPI path parameters",
								},
								"query": map[string]any{
									"type":        "object",
									"description": "Query parameters defined by the OpenAPI query parameters",
								},
								"header": map[string]any{
									"type":        "object",
									"description": "Header parameters defined by the OpenAPI header parameters",
								},
								"cookie": map[string]any{
									"type":        "object",
									"description": "Cookie parameters defined by the OpenAPI cookie parameters",
								},
								"body": map[string]any{
									"type":        "any",
									"description": "Request body parsed according to the OpenAPI request body schema",
								},
								"api": map[string]any{
									"type":        "string",
									"description": "Name of the API, as defined in the OpenAPI info.title field",
								},
							},
							"examples": []string{
								"request.body.userId",
								"request.query.page",
								"request.path.id",
								"request.key === '/users/{id}'",
							},
						},
						map[string]any{
							"name":        "response",
							"type":        "object",
							"description": "Used to define the outgoing HTTP response, including status code, headers, and response body.",
							"rules": []string{
								"Use response.data for structured responses (recommended)",
								"Use response.body only for raw responses",
								"Do not set both data and body at the same time",
							},
							"properties": map[string]any{
								"statusCode": map[string]any{
									"type":        "integer",
									"description": "HTTP status code used to select the OpenAPI response definition",
								},
								"headers": map[string]any{
									"type":        "object",
									"description": "Response headers defined by the OpenAPI response header parameters",
								},
								"body": map[string]any{
									"type":        "string",
									"description": "Raw response body. Takes precedence over data. Use body to return a raw response body without OpenAPI encoding and validating.",
								},
								"data": map[string]any{
									"type":        "any",
									"description": "Response body as data that will be encoded according to the OpenAPI response schema. This data must be valid against the response OpenAPI schema",
								},
							},
							"examples": []string{
								"response.data = { id: 1 }",
								"response.statusCode = 400",
							},
						},
						map[string]any{
							"name":        "args",
							"type":        "EventArgs",
							"description": "An optional configuration object. It allows controlling how and when an event handler is executed.",
						},
					},
					"bestPractices": []any{
						"Use switch(request.key) to route endpoints in HTTP event handling",
					},
					"useCases": []any{
						"Mock REST APIs",
						"Validate requests",
						"Simulate different responses",
					},
				},
				map[string]any{
					"name":        "on",
					"type":        "kafka",
					"package":     "mokapi",
					"signature":   "on('kafka', (message) => void, args?)",
					"description": "Handles incoming Kafka messages",
					"parameters": []any{
						map[string]any{
							"name":        "message",
							"type":        "KafkaEventMessage",
							"description": "Contains Kafka-specific message data.",
						},
						map[string]any{
							"name":        "args",
							"type":        "EventArgs",
							"description": "An optional configuration object. It allows controlling how and when an event handler is executed.",
						},
					},
				},
			},

			"scheduler": []any{
				map[string]any{
					"name":        "every",
					"signature":   "every(interval, handler, args?)",
					"description": "Runs a scheduled job (e.g. '5m', '10s', '1h30m')",
					"parameters": []any{
						map[string]any{
							"name":        "interval",
							"type":        "string",
							"description": "Valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'.",
						},
						map[string]any{
							"name":        "handler",
							"type":        "function",
							"description": "The handler function to be executed every interval. By default, the first execution happens immediately.",
						},
						map[string]any{
							"name":        "args",
							"type":        "ScheduledEventArgs",
							"description": "Contains additional event arguments.",
						},
					},
				},
			},
			"types": map[string]any{
				"EventArgs": map[string]any{
					"type":        "object",
					"description": "EventArgs is an optional configuration object passed to the on function when registering an event handler. It allows controlling how and when an event handler is executed.",
					"properties": map[string]any{
						"tags": map[string]any{
							"type":        "object",
							"description": "Adds or overrides existing tags that are used in dashboard.",
						},
						"track": map[string]any{
							"type":        "boolean | (params) => boolean",
							"description": "Controls whether this event handler is tracked in the dashboard. By default Mokapi resolves this based on whether the handler changes a parameter",
						},
						"priority": map[string]any{
							"type":        "integer",
							"description": "Defines the execution priority of the event handler. Handlers with a higher value are executed first. If no priority is specified, the default priority is 0.",
						},
					},
				},
				"ScheduledEventArgs": map[string]any{
					"type":        "object",
					"description": "ScheduledEventArgs is an object used by every and cron function.",
					"properties": map[string]any{
						"tags": map[string]any{
							"type":        "object",
							"description": "Adds or overrides existing tags that are used in dashboard.",
						},
						"times": map[string]any{
							"type":        "integer",
							"description": "How many times the job should execute (-1 for unlimited).",
							"default":     -1,
						},
						"skipImmediateFirstRun": map[string]any{
							"type":        "boolean",
							"description": "Toggles behavior of first execution. If true job does not start immediately but rather wait until the first scheduled interval.",
							"default":     false,
						},
					},
				},
				"KafkaEventMessage": map[string]any{
					"type":        "object",
					"description": "Kafka message including topic, key, value, headers",
					"properties": map[string]any{
						"offset": map[string]any{
							"type":        "integer",
							"description": "Offset of the kafka message",
						},
						"key": map[string]any{
							"type":        "string",
							"description": "The key of the message",
						},
						"value": map[string]any{
							"type":        "string",
							"description": "The value of the message",
						},
						"headers": map[string]any{
							"type":        "object",
							"description": "Kafka message headers",
						},
					},
				},
				"FetchOptions": map[string]any{
					"type":        "object",
					"description": "The FetchOptions object defines additional parameters for the fetch() function in the mokapi/http module.It allows you to customize request behavior such as HTTP method, headers, body, timeout, and redirect handling.",
					"properties": map[string]any{
						"method": map[string]any{
							"type":        "string",
							"description": "The HTTP method used for the request (e.g. 'GET', 'POST', 'PUT'). Defaults to 'GET'.",
						},
						"body": map[string]any{
							"type":        "any",
							"description": "The request body to send. Automatically serialized to JSON when Content-Type is set to application/json.",
						},
						"headers": map[string]any{
							"type":        "object",
							"description": "Key-value pairs representing HTTP headers to include with the request.",
						},
						"maxRedirects": map[string]any{
							"type":        "integer",
							"description": "The number of redirects to follow. Default value is 5. A value of 0 (zero) prevents all redirection.",
						},
						"timeout": map[string]any{
							"type":        "integer | string",
							"description": "Maximum time to wait for the request to complete. Default timeout is 60 seconds ('60s')",
						},
					},
				},
				"FetchResponse": map[string]any{
					"type":        "object",
					"description": "Response object contains response data from an HTTP request from methods in the module mokapi/http.",
					"properties": map[string]any{
						"body": map[string]any{
							"type":        "string",
							"description": "Response body content",
						},
						"headers": map[string]any{
							"type":        "object",
							"description": "All response headers sent by the server in canonical form (for example 'Content-Type'). Accessing a header value returns an array of strings.",
						},
						"statusCode": map[string]any{
							"type":        "integer",
							"description": "HTTP status code returned by the server",
						},
					},
					"functions": map[string]any{
						"json": map[string]any{
							"signature":   "json()",
							"description": "Parses the response body as JSON. Returns a JS object, array or value",
						},
					},
					"rules": []string{
						"Use json() when the response is JSON",
						"Use body for raw responses",
					},
				},
				"ProduceArgs": map[string]any{
					"type":        "object",
					"description": "ProduceArgs is an object used by produce function",
					"rules": []string{
						"Use message.data for structured message (recommended)",
						"Use message.value only for raw message",
						"Do not set both data and value at the same time",
					},
					"properties": map[string]any{
						"cluster": map[string]any{
							"type":        "string",
							"description": "The Kafka API name",
						},
						"topic": map[string]any{
							"type":        "string",
							"description": "The Kafka topic where the message is sent",
						},
						"message": map[string]any{
							"type":        "array",
							"description": "A list of Message contains Kafka messages to produce into given topic",
							"items": map[string]any{
								"type":        "object",
								"description": "Message represents a Kafka message used by produce function",
								"properties": map[string]any{
									"partition": map[string]any{
										"type":        "integer",
										"description": "Kafka partition index. If not specified, the message will be written to any partition",
										"optional":    true,
									},
									"key": map[string]any{
										"type":        "any",
										"description": "Kafka message key. If not specified, a random key will be generated based on the topic configuration.",
										"optional":    true,
									},
									"data": map[string]any{
										"type":        "any",
										"description": "Kafka message data that is validated against schema. If data and value are not specified, a random value will be generated based on the topic configuration. Object will be encoded based on the topic configuration",
										"optional":    true,
									},
									"value": map[string]any{
										"type":        "string | number | boolean",
										"description": "Kafka message value that is not validated against schema. If value and data are not specified, a random value will be generated based on the topic configuration. Object will be encoded based on the topic configuration",
										"optional":    true,
									},
									"headers": map[string]any{
										"type":        "object",
										"description": "Kafka message headers.",
										"optional":    true,
									},
								},
							},
						},
					},
				},
			},

			"utils": []any{
				map[string]any{
					"name":        "sleep",
					"package":     "mokapi",
					"signature":   "sleep(duration)",
					"description": "Suspends the execution for the specified duration.",
					"parameters": []any{
						map[string]any{
							"name":        "duration",
							"type":        "number | string",
							"description": "Duration in milliseconds or duration as string with unit.Valid time units are ns, us (or µs), ms, s, m, h",
						},
					},
					"useCases": []any{
						"Delay HTTP response",
					},
				},
				map[string]any{
					"name":        "fetch",
					"package":     "mokapi/http",
					"signature":   "fetch(url, options?)",
					"description": "The fetch() function provides a Promise-based interface for making HTTP requests in Mokapi’s JavaScript environment. It works similarly to the browser’s Fetch API.",
					"parameters": []any{
						map[string]any{
							"name":        "url",
							"type":        "string",
							"description": "Request URL",
						},
						map[string]any{
							"name":        "options",
							"type":        "FetchOptions",
							"description": "Contains additional request parameters",
						},
					},
					"returns": map[string]any{
						"type": "Promise<FetchResponse>",
					},
					"examples": []string{
						"const res = await fetch('https://api.example.com/users')",
						"const data = await res.json()",
					},
					"useCases": []any{
						"Invoke a callback request",
						"Collect data from an external API",
						"Validate and forward requests to a real backend service",
					},
				},
				map[string]any{
					"name":        "produce",
					"package":     "mokapi/kafka",
					"signature":   "produce(args?)",
					"description": "Sends a Kafka message to a Kafka topic.",
					"parameters": []any{
						map[string]any{
							"name":        "args",
							"type":        "ProduceArgs",
							"description": "Contains Kafka produce arguments",
						},
					},
					"returns": map[string]any{
						"type": "ProduceResult",
					},
					"examples": []string{
						`const result = produce({
        topic: 'topic', 
        messages: [
            { key: 'key1', value: 'hello Mokapi' },
            { key: 'key2', value: 'hello world' }
        ],
    })`,
					},
					"useCases": []any{
						"Produce a Kafka message to simulate real world events",
					},
				},
			},
		},
	}, nil
}
