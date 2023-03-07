import {metrics} from "metrics";

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
            type: "string",
            format: "int64"
        },
        name: {
            type: "string"
        }
    },
    xml: {name: "Tag"}
}

const pet = {
    type: "object",
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

export let apps = [
    {
        name: "Swagger Petstore",
        description: "This is a sample server Petstore server. You can find out more about at [http://swagger.io](http://swagger.io)",
        version: "1.0.6",
        contact: {
            name: "Swagger Petstore Team",
            email: "petstore@petstore.com"
        },
        metrics: metrics.filter(x => x.name.startsWith("http")),
        servers: [
            {
                url: "http://localhost:8080",
                description: "Server is mocked by *mokapi*"
            }
        ],
        paths: [
            {
                path: "/pet",
                operations: [
                    {
                        method: "post",
                        summary: "Add a new pet to the store",
                        operationId: "addPet",
                        requestBody: {
                            description: "Pet object that needs to be added to the store",
                            contents: [
                                {
                                    type: "application/json",
                                    schema: {
                                        type: "object",
                                        properties: {
                                            id: {
                                                type: "integer",
                                                format: "int64"
                                            }
                                        }
                                    }
                                }
                            ]
                        },
                        responses: [
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
                                }
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
            }
        ]
    }
]

export let events = [
    {
        id: "4242",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet"
        },
        time:  '2023-02-13T08:49:25.482366+01:00',
        data: {
            deprecated: true,
            request: {
                method: "POST",
                url: "http://127.0.0.1:18080/pet",
                parameters: [
                    {
                        name: 'Acceot-Encoding',
                        type: 'header',
                        raw: 'gzip, deflate'
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
            duration: 30016
        },
        workflows: [
            {
                duration: 20,
                tags: {
                    name: "dashboard",
                    file: "/Users/maesi/GolandProjects/mokapi/examples/mokapi/http_handler.js",
                    event: "http"
                }
            }
        ]
    },
    {
        id: "a5348a42-b69d-409f-8698-044d2ba845a3",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet/findByStatus"
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
            duration: 133
        },
        workflows: [
            {
                duration: 20,
                tags: {
                    name: "dashboard",
                    file: "/Users/maesi/GolandProjects/mokapi/examples/mokapi/http_handler.js",
                    event: "http"
                }
            }
        ]
    }
]