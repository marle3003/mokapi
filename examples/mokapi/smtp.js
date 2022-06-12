import {metrics} from "./metrics";

export let server = [
    {
        name: "Smtp Testserver",
        description: "This is a sample smtp server",
        version: "1.0",
        address: "localhost:8025",
        metrics: metrics.filter(x => x.name.startsWith("smtp"))
    }
]