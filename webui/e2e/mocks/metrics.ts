const app_start = new Date()
app_start.setHours(new Date().getHours() - 2)

export let metrics = [
    {
        name: 'app_start_timestamp',
        value:  1652025690 //app_start.getTime() / 1000,// 1652025690,
    },
    {
        name: 'app_memory_usage_bytes',
        value: 52450600
    },
    {
        name: 'app_job_run_total',
        value:  1
    },
    {
        name: 'http_requests_total{service="Swagger Petstore",endpoint="/pet",method="POST"}',
        value: 2
    },
    {
        name: 'http_requests_errors_total{service="Swagger Petstore",endpoint="/pet",method="POST"}',
        value: 1
    },
    {
        name: 'http_requests_total{service="Swagger Petstore",endpoint="/pet/findByStatus",method="GET"}',
        value: 2
    },
    {
        name: 'http_requests_errors_total{service="Swagger Petstore",endpoint="/pet/findByStatus",method="POST"}',
        value: 0
    },
    {
        name: 'http_request_timestamp{service="Swagger Petstore",endpoint="/pet",method="POST"}',
        value: 1652235690
    },
    {
        name: 'http_request_timestamp{service="Swagger Petstore",endpoint="/pet/findByStatus",method="POST"}',
        value: 1652237690
    },
    {
        name: 'kafka_messages_total{service="Kafka World",topic="mokapi.shop.products"}',
        value: 10
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="mokapi.shop.products"}',
        value: 1652135690
    },
    {
        name: 'kafka_consumer_group_lag{service="Kafka World",group="foo",topic="mokapi.shop.products",partition="0"}',
        value: 10
    },
    {
        name: 'kafka_consumer_group_commit{service="Kafka World",group="foo",topic="mokapi.shop.products",partition="0"}',
        value: 1
    },
    {
        name: 'kafka_rebalance_timestamp{service="Kafka World",group="foo"}',
        value: 1652135690
    },
    {
        name: 'kafka_messages_total{service="Kafka World",topic="bar"}',
        value: 1
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="bar"}',
        value: 0
    },
    {
        name: 'kafka_messages_total{service="Kafka World",topic="mokapi.shop.userSignedUp"}',
        value: 0
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="mokapi.shop.userSignedUp"}',
        value: 0
    },
    {
        name: 'kafka_messages_total{service="Kafka World",topic="mokapi.shop.avro"}',
        value: 0
    },
    {
        name: 'kafka_message_timestamp{service="Kafka World",topic="mokapi.shop.avro"}',
        value: 1652035690
    },
    {
        name: 'mail_mails_total{service="Mail Testserver"}',
        value: 3
    },
    {
        name: 'mail_mail_timestamp{service="Mail Testserver"}',
        value: 1652635690
    },
    {
        name: 'ldap_requests_total{service="LDAP Testserver"}',
        value: 6
    },
    {
        name: 'ldap_request_timestamp{service="LDAP Testserver"}',
        value: 1652635690
    },
    // mqtt
    {
        name: 'mqtt_messages_total{service="MQTT Temperature Sensor API",topic="home/livingroom/temperature"}',
        value: 1
    },
    {
        name: 'mqtt_message_timestamp{service="MQTT Temperature Sensor API",topic="home/livingroom/temperature"}',
        value: 1672182690
    },
    {
        name: 'mqtt_messages_total{service="MQTT Temperature Sensor API",topic="sensors/{sensorId}/data"}',
        value: 1
    },
    {
        name: 'mqtt_message_timestamp{service="MQTT Temperature Sensor API",topic="sensors/{sensorId}/data"}',
        value: 1771058965
    },
]

export function sum(name, ...labels){
    if (!metrics){
        return 0
    }

    let sum = 0
    for (let metric of metrics) {
        if (!metric.name.startsWith(name)) {
            continue
        }
        
        if (labels.length == 0 || matchLabels(metric, labels)){
            sum += Number(metric.value)
        }
    }  
    return sum
}

export function max(name, ...labels) {
    if (!metrics){
        return 0
    }

    let max = 0
    for (let metric of metrics) {
        if (!metric.name.startsWith(name)) {
            continue
        }
        
        if (labels.length == 0 || matchLabels(metric, labels)){
            const n = Number(metric.value)
            if (n > max) {
                max = n
            }
        }
    }  
    return max
}

export function filter(name, ...labels) {
    const result = [];
    if (!metrics){
        return [];
    }

    for (let metric of metrics) {
        if (!metric.name.startsWith(name)) {
            continue
        }
        
        if (labels.length == 0 || matchLabels(metric, labels)){
            result.push(metric)
        }
    }  
    return result;
}

export function parseLabels(metric) {
    const match = metric.name.match(/\{(.+)}/);
    if (!match || match.length < 2 || !match[1]) {
        return {};
    }

    return Object.fromEntries(
        Array.from(
            match[1].matchAll(/(\w+)="([^"]*)"/g),
            m => [m[1], m[2]]
        )
    );
}

function matchLabels(metric, labels){
    for (var label of labels){
        const s = `${label.name}="${label.value}"`
        if (!metric.name.includes(s)){
            return false
        }
    }
    return true
}
