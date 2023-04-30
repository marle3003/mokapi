import {metrics} from "./metrics.js";

export let server = [
    {
        name: "Smtp Testserver",
        description: "This is a sample smtp server",
        version: "1.0",
        server: "smtp://localhost:8025",
        mailboxes: [
            {name: 'alice@mokapi.io', username: 'alice', password: 'foo'},
            {name: 'bob@mokapi.io', username: 'bob', password: 'foo'},
        ],
        rules: [
            {sender: '.*@mokapi.io', action: 'allow'}
        ],
        metrics: metrics.filter(x => x.name.startsWith("smtp"))
    }
]

export let mails = [
    {
        id: "8832",
        traits: {
            namespace: "smtp",
            name: "Smtp Testserver",
        },
        time:  '2023-02-23T08:49:25.482366+01:00',
        data: {
            from: 'alice@mokapi.io',
            to: ['bob@mokapi.io'],
            mail: {
                From: {Name: 'Alice', Address: 'alice@mokapi.io'},
                To: [{Name: 'Bob', Address: 'bob@mokapi.io'}],
                Date: '2023-02-23T08:49:25.482366+01:00',
                Subject: 'A test mail',
                Body: 'Mail message from Alice'
            },
            duration: 30016,
            actions: [
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
    },
]