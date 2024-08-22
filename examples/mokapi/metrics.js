const app_start = new Date()
app_start.setHours(new Date().getHours() - 2)

export let metrics = [
    {
        name: 'app_start_timestamp',
        value: app_start.getTime() / 1000,// 1652025690,
    },
    {
        name: 'app_memory_usage_bytes',
        value: 52450600
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
        value: 2
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
        name: 'kafka_messages_total{service="Kafka World",topic="mokapi.shop.products"}"',
        value: 10
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="mokapi.shop.products"}"',
        value: 1652135690
    },
    {
        name: 'kafka_consumer_group_lag{service="Kafka World",group="foo",topic="mokapi.shop.products",partition="0"}"',
        value: 10
    },
    {
        name: 'kafka_messages_total{service="Kafka World",topic="bar"}"',
        value: 1
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="bar"}"',
        value: 1652035690
    },
    {
        name: 'smtp_mails_total{service="Smtp Testserver"}"',
        value: 3
    },
    {
        name: 'smtp_mail_timestamp{service="Smtp Testserver"}"',
        value: 1652635690
    },
    {
        name: 'ldap_search_total{service="LDAP Testserver"}"',
        value: 1
    },
    {
        name: 'ldap_search_timestamp{service="LDAP Testserver"}"',
        value: 1652635690
    },
]