import {metrics} from "metrics";

export let apps = [
    {
        name: "Swagger Petstore",
        description: "This is a sample server Petstore server",
        version: "1.0.6",
        metrics: metrics.filter(x => x.name.startsWith("http"))
    }
]

export let events = [
    {
        id: "4242",
        traits: {
            namespace: "http",
        },
        time: 1651771269,
        data: {
            request: {
                method: "GET",
                url: "http://127.0.0.1:18080/demo",
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
        },
        time: 1654771269,
        data: {
            request: {
                method: "POST",
                url: "http://127.0.0.1:18080/demo",
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