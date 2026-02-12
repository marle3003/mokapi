import {metrics} from "metrics"

const category = {
    type: "object",
    properties: {
        id: {
            type: "integer",
            format: "int64",
        },
        name: {
            type: "string"
        }
    },
    xml: {name: "Category"}
}

const tag = {
    type: "object",
    properties: {
        id: {
            type: "integer",
            format: "int64"
        },
        name: {
            type: "string"
        }
    },
    xml: {name: "Tag"}
}

const pet = {
    type: ["object"],
    properties: {
        id: {
            type: "integer",
            format: "int64",
        },
        category: {
            ref: "#/components/schemas/Category",
            ...category
        },
        name: {
            type: "string",
            example: "doggie"
        },
        photoUrls: {
            type: "array",
            xml: {
                name: "photoUrl",
                wrapped: true
            },
            items: {
                type: "string"
            }
        },
        tags: {
            type: "array",
            xml: {
                name: "tag",
                wrapped: true
            },
            items: {
                ref: "#/components/schemas/Tag",
                ...tag
            }
        },
        status: {
            type: "string",
            description: "pet status in the store",
            enum: ["available", "pending", "sold"]
        }
    },
    xml: {name: "Pet"}
}

const book = {
    type: "object",
    properties: {
        id: {
            type: "string",
            example: 1
        },
        title: {
            type: "string",
            example: "The Hobbit"
        },
        author: {
            type: "string",
            example: "J.R.R. Tolkien"
        },
        year: {
            type: "integer",
            example: 1937
        }
    },
    required: ["id", "title", "author"]
}

export const configs = {
    'b6fea8ac-56c7-4e73-a9c0-6887640bdca8': {
        id: 'b6fea8ac-56c7-4e73-a9c0-6887640bdca8',
        url: 'file://foo.json',
        provider: 'file',
        time: '2023-02-15T08:49:25.482366+01:00',
        data: 'http://localhost:8090/api/services/http/Swagger%20Petstore',
        refs: [
            {
                id: 'b6fea8ac-56c7-4e73-a9c0-6887640bdba8',
                url: 'file://foo2.json',
                provider: 'file',
                time: '2023-02-28T08:49:25.482366+01:00',
                data: 'http://localhost:8090/api/services/http/Swagger%20Petstore',
            }
        ],
        filename: 'petstore.json',
        tags: ['Tag1', 'Tag2'],
    },
    'b6fea8ac-56c7-4e73-a9c0-4487640bdca8': {
        id: 'b6fea8ac-56c7-4e73-a9c0-4487640bdca8',
        url: 'https://github.com/marle3003/mokapi.git?file=foo.json&ref=main',
        provider: 'git',
        time: '2023-02-15T08:49:25.482366+01:00',
        data: 'http://localhost:8090/api/services/http/Swagger%20Petstore',
        refs: [],
        filename: 'petstore.json'
    }
}

export let apps = [
    {
        name: "Swagger Petstore",
        description: "This is a sample server Petstore server. You can find out more about at [http://swagger.io](http://swagger.io)",
        version: "1.0.6",
        contact: {
            name: "Swagger Petstore Team",
            email: "petstore@petstore.com"
        },
        metrics: metrics.filter(x => x.name.startsWith("http") && x.name.includes('service="Swagger Petstore"')),
        servers: [
            {
                url: "http://localhost:8080",
                description: "Server is mocked by *mokapi*"
            }
        ],
        paths: [
            {
                path: "/pet",
                summary: "Everything about your Pets",
                operations: [
                    {
                        method: "post",
                        summary: "Add a new pet to the store",
                        operationId: "addPet",
                        deprecated: true,
                        tags: ["pet", "post"],
                        parameters: [
                            {
                                type: "query",
                                name: "$format",
                                schema: {type: "string"},
                                required: false,
                                deprecated: false,
                                explode: false,
                                allowReserved: false
                            }
                        ],
                        requestBody: {
                            description: "Create a new pet in the store",
                            required: true,
                            contents: [
                                {
                                    type: "application/json",
                                    schema: pet
                                },
                                {
                                    type: "application/xml",
                                    schema: pet
                                },
                                {
                                    type: "application/yaml",
                                    schema: pet
                                }
                            ]
                        },
                        responses: [
                            {
                                statusCode: 200,
                                description: "Successful operation",
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: pet
                                    }
                                ]
                            },
                            {
                                statusCode: 400,
                                description: "Invalid ID supplied",
                                headers: [
                                    {
                                        name: "petId",
                                        description: "Status values that need to be considered for filter",
                                        schema: {
                                            type: "number"
                                        }
                                    }
                                ]
                            }
                        ],
                        security: [
                            {
                                foo: {
                                    scopes: [],
                                    configs: {
                                        type: "apiKey",
                                        in: "header",
                                        name: "X-API-Key"
                                    }
                                },
                            },
                            {
                                foo: {
                                    scopes: [],
                                    configs: {
                                        type: "apiKey",
                                        in: "header",
                                        name: "X-API-Key"
                                    }
                                },
                                bar: {
                                    scopes: ["read_pets"],
                                    configs: {
                                        type: "oauth2",
                                        flows: {
                                            implicit: {
                                                authorizationUrl: "https://api.example.com/oauth2/authorize",
                                                scopes: {
                                                    read_pets: "read your pets",
                                                    write_pets: "modify pets in your account",
                                                }
                                            }
                                        }
                                    }
                                }
                            },
                            {
                                basicAuth: {
                                    scopes: [],
                                    configs: {
                                        type: "http",
                                        scheme: "basic"
                                    }
                                }
                            },
                            {
                                bearerAuth: {
                                    scopes: [],
                                    configs: {
                                        type: "http",
                                        scheme: "bearer",
                                        bearerFormat: "JWT"
                                    }
                                }
                            }
                        ]
                    }
                ]
            },
            {
                path: "/pet/{petId}",
                operations: [
                    {
                        method: "post",
                        summary: "Updates a pet in the store with form data",
                        tags: ["pet", "post"],
                        parameters: [
                            {
                                name: "petId",
                                type: "path",
                                description: "ID of pet that needs to be updated",
                                schema: {
                                    type: "integer",
                                    format: "int64"
                                },
                                required: false,
                                deprecated: false,
                                explode: false,
                                allowReserved: false
                            },
                            {
                                name: "name",
                                type: "formData",
                                description: "Updated name of the pet",
                                schema: {
                                    type: "string"
                                },
                                required: false,
                                deprecated: false,
                                explode: false,
                                allowReserved: false
                            },
                            {
                                name: "status",
                                type: "formData",
                                description: "Updated status of the pet",
                                schema: {
                                    type: "string"
                                },
                                required: false,
                                deprecated: false,
                                explode: false,
                                allowReserved: false
                            }
                        ],
                        responses: [
                            {
                                statusCode: 200,
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: {
                                            type: "object",
                                            properties: {
                                                id: {
                                                    type: "integer",
                                                    format: "int64"
                                                },
                                                name: {
                                                    type: "string"
                                                },
                                                status: {
                                                    type: "string"
                                                }
                                            }
                                        }
                                    }
                                ]
                            }
                        ]
                    },
                    {
                        method: "get",
                        summary: "Returns a single pet",
                        tags: ["pet"],
                        parameters: [
                            {
                                name: "petId",
                                type: "path",
                                description: "ID of pet to return",
                                schema: {
                                    type: "integer",
                                    format: "int64"
                                },
                                required: false,
                                deprecated: false,
                                explode: false,
                                allowReserved: false
                            }
                        ],
                        responses: [
                            {
                                statusCode: 200,
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: pet
                                    }
                                ]
                            }
                        ]
                    }
                ]
            },
            {
                path: "/pet/findByStatus",
                operations: [
                    {
                        method: "get",
                        summary: "Finds Pets by status",
                        description: "Multiple status values can be provided with comma separated strings",
                        operationId: "findPetsByStatus",
                        tags: ["pet"],
                        parameters: [
                            {
                                name: "status",
                                type: "query",
                                description: "Status values that need to be considered for filter",
                                required: true,
                                style: "form",
                                explode: true,
                                schema: {
                                    type: "array",
                                    items: {
                                        type: "string",
                                        enum: ["available", "pending", "sold"]
                                    }
                                },
                                deprecated: false,
                                allowReserved: false
                            }
                        ],
                        responses: [
                            {
                                statusCode: 200,
                                description: "successful operation",
                                headers: [
                                    {
                                        name: "Expires-After",
                                        schema: {type: "string", format: "date-time"},
                                        description: "date in UTC when token expires",
                                    }
                                ],
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: {
                                            type: "array",
                                            items: {
                                                ref: "#/components/schemas/Pet",
                                                ...pet
                                            }
                                        }
                                    },
                                    {
                                        type: "application/xml",
                                        schema: {
                                            type: "array",
                                            items: {
                                                ref: "#/components/schemas/Pet",
                                                ...pet
                                            },
                                            xml: {
                                                wrapped: true
                                            }
                                        }
                                    }
                                ]
                            },
                            {
                                statusCode: 500,
                                description: "server error",
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: {
                                            type: "object",
                                            properties: {
                                                error: {
                                                    type: "number"
                                                },
                                                message: {
                                                    type: "string"
                                                }
                                            }
                                        }
                                    }
                                ]
                            }
                        ]
                    }
                ]
            },
            {
                path: "/Zorem/ipsum/dolor/sit/amet/consetetur/sadipscing/elitr/sed/diam/nonumy/eirmod",
                summary: "A long path example",
                operations: [
                    {
                        method: "post",
                        summary: "A long path example",
                        tags: ["custom", "post"],
                        requestBody: {
                            description: "Create a new pet in the store",
                            contents: [
                                {
                                    type: "application/json",
                                    schema: pet
                                }
                            ]
                        },
                        responses: [
                            {
                                statusCode: 200,
                                description: "Successful operation",
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: pet
                                    }
                                ]
                            },
                            {
                                statusCode: 400,
                                description: "Invalid ID supplied",
                                headers: [
                                    {
                                        name: "petId",
                                        description: "Status values that need to be considered for filter",
                                        schema: {
                                            type: "number"
                                        }
                                    }
                                ]
                            },
                            {
                                statusCode: 'default',
                                description: 'Unexpected error'
                            }
                        ]
                    }
                ]
            }
        ],
        tags: [
            {
                name: "pet",
                summary: "Everything about your Pets"
            },
            {
                name: "custom",
                summary: "Access to Petstore orders"
            }
        ],
        configs: [
            configs['b6fea8ac-56c7-4e73-a9c0-6887640bdca8']
        ]
    },
    {
        name: "Books API",
        description: "A simple API to manage books in a library",
        version: "1.0.0",
        metrics: metrics.filter(x => x.name.startsWith("http") && x.name.includes('service="Books API"')),
        servers: [
            {
                url: "https://api.example.com/v1",
            }
        ],
        paths: [
            {
                path: "/books",
                operations: [
                    {
                        method: "get",
                        summary: "Get books from the store",
                        operationId: "listBooks",
                        responses: [
                            {
                                statusCode: 200,
                                description: "A list of books",
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: {
                                            type: "array",
                                            items: book
                                        }
                                    }
                                ]
                            },
                        ],
                    },
                    {
                        method: "post",
                        summary: "Add a new book",
                        operationId: "addBook",
                        requestBody: {
                            required: true,
                            contents: [
                                {
                                    type: "application/json",
                                    schema: book
                                }
                            ]
                        },
                        responses: [
                            {
                                statusCode: 201,
                                description: "The created book",
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: book
                                    }
                                ]
                            },
                        ],
                    }
                ]
            },
        ]
    }
]

export let events = [
    {
        id: "9e058601-562a-4063-a3b2-066ba624893a",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet",
            method: 'POST'
        },
        time:  '2023-02-13T08:49:25.482366+01:00',
        data: {
            deprecated: true,
            request: {
                method: "POST",
                url: "http://127.0.0.1:18080/pet",
                parameters: [
                    {
                        name: 'Accept-Encoding',
                        type: 'header',
                        raw: 'gzip, deflate'
                    },
                    {
                        name: 'LongHeader',
                        type: 'header',
                        raw: 'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.'
                    }
                ],
                contentType: "application/xml",
                body: "<foo bar=test><child><key>foo</key><value>bar</value></child>"
            },
            response: {
                statusCode: 200,
                headers: {
                    "Content-Type": "application/json"
                },
                body: "{\"foo\":\"bar\"}"
            },
            duration: 30016,
            actions: [
                {
                    duration: 20,
                    tags: {
                        name: "dashboard",
                        file: "examples/mokapi/http_handler.js",
                        fileKey: "b6fea8ac-56c7-4e73-a9c0-6887640bdca8",
                        event: "http"
                    },
                    logs: [
                        { level: 'info', message: 'a custom log message, see https://mokapi.io'},
                        { level: 'warn', message: 'a warn log message'},
                        { level: 'error', message: 'a error log message'},
                        { level: 'debug', message: 'a debug log message'}
                    ],
                    parameters: [
                        JSON.stringify({
                            method: "POST",
                            url: "http://127.0.0.1:18080/pet",
                            parameters: [
                                {
                                    name: 'Accept-Encoding',
                                    type: 'header',
                                    raw: 'gzip, deflate'
                                }
                            ],
                            contentType: "application/xml",
                            body: "<foo bar=test><child><key>foo</key><value>bar</value></child>"
                        }),
                        JSON.stringify({
                            statusCode: 200,
                            headers: {
                                "Content-Type": "application/json"
                            },
                            data: {"foo":"bar"}
                        })
                    ]
                }
            ],
            clientIP: '127.0.0.1'
        },
    },
    {
        id: "a5348a42-b69d-409f-8698-044d2ba845a3",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet/findByStatus",
            method: 'GET'
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            request: {
                method: "GET",
                url: "http://127.0.0.1:18080/pet/findByStatus",
                contentType: "application/json",
                body: "{\"foo\":\"bar\",\"child\":{\"key\":\"val\"}}",
                parameters: [
                    {
                        name: "foo",
                        type: "query",
                        value: "bar",
                        raw: "bar"
                    }
                ]
            },
            response: {
                statusCode: 201,
                headers: {
                    "Content-Type": "application/json",
                    "Set-Cookie": "sessionId=38afes7a8"
                },
                size: 512
            },
            duration: 133,
            actions: [
                {
                    duration: 20,
                    tags: {
                        name: "dashboard",
                        file: "examples/mokapi/http_handler.js",
                        fileKey: "b6fea8ac-56c7-4e73-a9c0-4487640bdca8",
                        event: "http"
                    },
                    error: {
                        message: 'An example script error message'
                    }
                }
            ],
            clientIP: '127.0.0.1'
        }
    },
    {
        id: "a5348a42-b69d-409f-8600-044d2ba845a3",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet/findByStatus",
            method: 'GET'
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            request: {
                method: "GET",
                url: "http://127.0.0.1:18080/pet/findByStatus",
                parameters: [
                    {
                        name: "foo",
                        type: "query",
                        value: "bar",
                        raw: "bar"
                    }
                ]
            },
            response: {
                statusCode: 201,
                headers: {
                    "Content-Type": "application/json",
                    "Set-Cookie": "sessionId=38afes7a8"
                },
                size: 512
            },
            duration: 133,
            actions: [
                {
                    duration: 20,
                    tags: {
                        name: "dashboard",
                        file: "examples/mokapi/http_handler.js",
                        fileKey: "b6fea8ac-56c7-4e73-a9c0-4487640bdca8",
                        event: "http"
                    }
                }
            ],
            clientIP: '127.0.0.1'
        }
    },
    {
        id: "ac509b4f-9254-4bb7-abbb-90c100310ad7",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet",
            method: 'POST'
        },
        time: '2023-02-13T10:05:25.583366+01:00',
        data: {
            request: {
                method: "POST",
                url: "http://127.0.0.1:18080/pet",
                contentType: "application/json",
                body: `{"id": 0,"category": {"id": 0,"name": "string"},"name": "doggie","photoUrls": ["string"],"tags":[{"id": 0,"name": "string"}],"status": "available"}`,
            },
            response: {
                statusCode: 500,
                headers: {
                    "Content-Type": "application/json",
                },
                body: `{"id": 0,"category": {"id": 0,"name": "string"},"name": "doggie","photoUrls": ["string"],"tags":[{"id": 0,"name": "string"}],"status": "available"}`,
                size: 512
            },
            duration: 133,
            actions: [
                {
                    duration: 20,
                    tags: {
                        name: "pet store",
                        file: "examples/mokapi/http_handler.js",
                        fileKey: "b6fea8ac-56c7-4e73-a9c0-6887640bdca8",
                        event: "http"
                    }
                }
            ],
            clientIP: '127.0.0.1'
        }
    },
    {
        id: "a5409fc1-6a5a-4c0b-be39-a6b80885e41b",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/requestWithParams",
            method: 'GET'
        },
        time: '2023-02-13T10:05:25.583366+01:00',
        data: {
            request: {
                method: "GET",
                url: "http://127.0.0.1:18080/requestWithParams",
                contentType: "application/json",
                parameters: [
                    {
                        name: "foo",
                        type: "query",
                        value: `"bar"`,
                        raw: "bar"
                    },
                    {
                        name: "debug",
                        type: "header",
                        value: `{"foo":"bar"}`,
                        raw: "bar"
                    },
                    {
                        name: "noSpec",
                        type: "header",
                        raw: "bar"
                    }
                ]
            },
            response: {
                statusCode: 500,
                headers: {
                    "Content-Type": "application/json",
                },
                body: `{"id": 0,"category": {"id": 0,"name": "string"},"name": "doggie","photoUrls": ["string"],"tags":[{"id": 0,"name": "string"}],"status": "available"}`,
                size: 512
            },
            duration: 133,
            clientIP: '192.0.1.127'
        }
    }
]