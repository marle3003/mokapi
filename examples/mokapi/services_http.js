import {metrics} from "metrics";

export let apps = [
    {
        name: "Swagger Petstore",
        description: "This is a sample server Petstore server",
        version: "1.0.6",
        metrics: metrics.filter(x => x.name.startsWith("http"))
    }
]