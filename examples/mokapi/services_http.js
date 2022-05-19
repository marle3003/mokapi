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
                body: "{foo: \"bar\"}"
            },
            response: {
                statusCode: 200,
                Headers: {
                    "Content-Type": "application/json"
                },
                body: "{foo: \"bar\"}"
            }
        }
    }
]