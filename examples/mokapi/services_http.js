import {metrics} from "metrics";

export let apps = [
    {
        name: "Swagger Petstore",
        description: "This is a sample server Petstore server",
        version: "1.0.6",
        metrics: metrics.filter(x => x.name.startsWith("http")),
        servers: [
            {
                url: "http://localhost:8080"
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
                                description: "Invalid ID supplied"
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
                                contents: [
                                    {
                                        type: "application/json",
                                        schema: {
                                            type: "array",
                                            items: {
                                                ref: "#/components/schemas/Pet"
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
        time: 1651771269,
        data: {
            request: {
                method: "GET",
                url: "http://127.0.0.1:18080/pet",
                parameters: [],
                contentType: "application/json",
            },
            response: {
                statusCode: 200,
                headers: {
                    "Content-Type": "application/json"
                },
                body: "{\"foo\":\"bar\"}"
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
    },
    {
        id: "4243",
        traits: {
            namespace: "http",
            name: "Swagger Petstore",
            path: "/pet/findByStatus"
        },
        time: 1654771269,
        data: {
            request: {
                method: "POST",
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