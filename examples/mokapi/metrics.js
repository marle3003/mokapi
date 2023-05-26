export let metrics = [
    {
        name: 'app_start_timestamp',
        value: 1652025690,
    },
    {
        name: 'app_memory_usage_bytes',
        value: 124506000
    },
    {
        name: 'http_requests_total{service="Swagger Petstore",endpoint="/pet"}"',
        value: 2
    },
    {
        name: 'http_requests_errors_total{service="Swagger Petstore",endpoint="/pet"}"',
        value: 1
    },
    {
        name: 'http_requests_total{service="Swagger Petstore",endpoint="/pet/findByStatus"}"',
        value: 1
    },
    {
        name: 'http_requests_errors_total{service="Swagger Petstore",endpoint="/pet/findByStatus"}"',
        value: 0
    },
    {
        name: 'http_request_timestamp{service="Swagger Petstore,endpoint="/pet"}"',
        value: 1652235690
    },
    {
        name: 'http_request_timestamp{service="Swagger Petstore,endpoint="/pet/findByStatus"}"',
        value: 1652237690
    },
    {
        name: 'kafka_messages_total{service="Kafka World",topic="foo"}"',
        value: 10
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="foo"}"',
        value: 1652135690
    },
    {
        name: 'kafka_consumer_group_lag{service="Kafka World",group="foo",topic="foo",partition="0"}"',
        value: 10
    },
    {
        name: 'smtp_mails_total{service="Smtp Testserver"}"',
        value: 3
    },
    {
        name: 'smtp_mail_timestamp{service="Smtp Testserver"}"',
        value: 1652635690
    },
]