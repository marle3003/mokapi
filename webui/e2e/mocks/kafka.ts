import { app } from 'mokapi'
import { sum, max, metrics, filter as filterMetrics, parseLabels } from './metrics'
import { base64 } from 'mokapi/encoding';
import { getInfo, matchTraits } from './helpers';

export default function () {
    app.http()
        .operation('serviceKafka', (_, res) => {
            res.data = {
                name: clusters[0].name,
                description: clusters[0].description,
                version: clusters[0].version,
                contact: clusters[0].contact,
                servers: clusters[0].servers,
                topics: clusters[0].topics.filter(x => !x.template).map(x => {
                    return {
                        name: x.name,
                        ...x.summary === undefined ? {} : { summary: x.summary },
                        ...('parameters' in x ? { parameters: x.parameters } : {}),
                        ...('instances' in x ? { instances: x.instances?.map((i: any) => (
                            {
                                ...i,
                                metrics: {
                                messages_total: sum('kafka_messages_total', { name: 'service', value: clusters[0].name }, { name: 'topic', value: x.name }, { name: 'instance', value: i.name }),
                                message_timestamp: max('kafka_message_timestamp', { name: 'service', value: clusters[0].name }, { name: 'topic', value: x.name }, { name: 'instance', value: i.name })
                                },
                            }
                            )) } : {}),
                        metrics: {
                            messages_total: metrics.find(m => m.name === `kafka_messages_total{service="Kafka World",topic="${x.name}"}`)?.value ?? 0,
                            message_timestamp: metrics.find(m => m.name === `kafka_message_timestamp{service="Kafka World",topic="${x.name}"}`)?.value ?? 0
                        }
                    }
                }),
                groups: clusters[0].groups.map(x => {
                    return {
                        name: x.name,
                        generation: x.generation,
                        state: x.state,
                        protocol: x.protocol,
                        members: x.members.length ?? 0,
                        metrics: x.metrics
                    }
                }),
                clients: clusters[0].clients.map(c => {
                    return {
                        clientId: c.clientId,
                        address: c.address,
                        software: c.software,
                    }
                }),
                configs: clusters[0].configs.map(c => {
                    return {
                        id: c.id,
                        url: c.url,
                        provider: c.provider,
                        time: c.time
                    }
                })
            }
        })
        .operation('kafkaGroup', (req, res) => {
            res.data = clusters[0].groups.find(x => x.name === req.path['group'])
        })
        .operation('kafkaClusters', (req, res) => {
            res.data = clusters.map(x => getInfo(x, 'kafka'))
        })
        .operation('kafkaTopics', (req, res) => {
            res.data = clusters[0].topics
        })
        .operation('kafkaTopic', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic) {
                res.rebuild(404)
                return
            }
            const { parameters, ...data } = topic
            res.data = {
                ...data,
                ...('parameters' in topic && topic.template ? { parameters: topic.parameters }: {}),
                groups: clusters[0].groups.filter(g => g.topics.includes(topic.name)).map(x => {
                    return {
                        name: x.name,
                        generation: x.generation,
                        state: x.state,
                        protocol: x.protocol,
                        members: x.members.length ?? 0,
                        metrics: x.metrics
                    }
                }),
            }
        })
        .operation('kafkaProduceTopic', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic) {
                res.rebuild(404)
                return
            }
            res.data = {
                offsets: [
                    {
                        partition: 0,
                        offset: 1234,
                    }
                ]
            }
        })
        .operation('kafkaPartitions', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic) {
                res.rebuild(404)
                return
            }
            res.data = topic.partitions
        })
        .operation('kafkaPartition', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic || !topic.partitions) {
                res.rebuild(404)
                return
            }
            const id = req.path.partitionId
            if (id >= topic.partitions.length) {
                res.rebuild(404)
                return
            }
            res.data = topic.partitions[id]
        })
        .operation('kafkaProducePartition', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic || !topic.partitions) {
                res.rebuild(404)
                return
            }
            const id = req.path.partitionId
            if (id >= topic.partitions.length) {
                res.rebuild(404)
                return
            }
            res.data = {
                offsets: [
                    {
                        partition: id,
                        offset: 1234,
                    }
                ]
            }
        })
        .operation('kafkaOffsets', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic || !topic.partitions) {
                res.rebuild(404)
                return
            }
            const id = req.path.partitionId
            if (id >= topic.partitions.length) {
                res.rebuild(404)
                return
            }
            const data = events.filter(x => matchTraits(x, { topic: topic.name, type: 'message' }) && x.data.partition === id)
            if (req.header['Accept'] === 'application/json') {
                res.data = data.map(x => ({
                    key: x.data.key?.value,
                    value: JSON.parse(x.data.message?.value ?? '')
                }))
            } else {
                res.data = data.map(x => ({
                    key: base64.encode(x.data.key?.value ?? ''),
                    value: base64.encode(x.data.message?.value ?? '')
                }))
            }
        })
        .operation('kafkaOffset', (req, res) => {
            const topic = clusters[0].topics.find(x => x.name === req.path.topic)
            if (!topic || !topic.partitions) {
                res.rebuild(404)
                return
            }
            const id = req.path.partitionId
            if (id >= topic.partitions.length) {
                res.rebuild(404)
                return
            }
            const data = events.filter(x => matchTraits(x, { topic: topic.name, type: 'message' }) && x.data.partition === id && x.data.offset === req.path.offset)
            if (data.length === 0) {
                res.rebuild(404)
                return
            }
            if (req.header['Accept'] === 'application/json') {
                res.data = {
                    key: events[0].data.key?.value,
                    value: JSON.parse(events[0].data.message?.value ?? '')
                }
            } else {
                res.data = {
                    key: base64.encode(data[0].data.key?.value ?? ''),
                    value: base64.encode(data[0].data.message?.value ?? '')
                }
            }
        })
        .operation('kafkaClient', (req, res) => {
            res.data = clusters[0].clients.find(x => x.clientId === req.path['clientId'])
        })
}

const Product = {
    type: 'object',
    properties: {
        id: {
            type: 'string',
        },
        name: {
            type: 'string'
        },
        description: {
            type: 'string'
        },
        features: {
            type: 'string'
        },
        price: {
            type: 'number'
        },
        keywords: {
            type: 'string'
        },
        url: {
            type: 'string'
        },
        category: {
            type: 'string'
        },
        subcategory: {
            type: 'string'
        }
    }
}

const Order = {
    type: 'object',
    properties: {
        orderId: { type: 'string', format: 'uuid' },
        amount: { type: 'number' },
        timestamp: { type: 'string', format: 'date-time' }
    }
}

const UserSignedUp = {
    type: 'object',
    properties: {
        userId: {
            type: 'integer',
            minimum: 1,
            description: 'This property describes the id of the user'
        },
        userEmail: {
            type: 'string',
            format: 'email',
            description: 'This property describes the email of the user'
        }
    },
    required: ['userId', 'userEmail'],
    xml: {
        name: 'user'
    }
}

const SecondMessage = {
    type: 'object',
    properties: {
        foo: {
            type: 'string'
        },
        bar: {
            type: 'integer'
        }
    },
}


export const configs = {
    'b6fea8ac-56c7-4e73-a9c0-6337640bdca8': {
        id: 'b6fea8ac-56c7-4e73-a9c0-6337640bdca8',
        url: 'https://www.example.com/foo/bar/communication/service/asyncapi.json',
        provider: 'http',
        time: '2023-02-15T08:49:25.482366+01:00',
        data: 'http://localhost:8090/api/services/kafka/Kafka%20World',
        filename: 'asyncapi.json'
    }
}

export let clusters = [
    {
        name: 'Kafka World',
        description: 'To ours significant why upon tomorrow her faithful many motionless.',
        version: '4.01',
        contact: {
            name: 'mokapi',
            url: 'https://www.mokapi.io',
            email: 'info@mokapi.io'
        },
        servers: [
            {
                name: 'Broker',
                host: 'localhost:9092',
                tags: [{ name: 'env:test', description: 'This environment is for running internal tests' }],
                description: 'Dashwood contempt on mr unlocked resolved provided of of. Stanhill wondered it it welcomed oh. Hundred no prudent he however smiling at an offence. If earnestly extremity he he propriety something admitting convinced ye.',
                configs: {
                    'log.retention.bytes': 1024,
                    'log.retention': 5000,
                }
            }
        ],
        topics: [
            {
                name: 'mokapi.shop.products',
                summary: 'Though literature second anywhere fortnightly am this either so me.',
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 4,
                        leader: { name: 'foo', addr: 'localhost:9002' },
                        segments: 1
                    },
                    {
                        id: 1,
                        startOffset: 0,
                        offset: 3,
                        leader: { 'name': 'foo', 'addr': 'localhost:9002' },
                        segments: 1
                    },
                    {
                        id: 2,
                        startOffset: 0,
                        offset: 3,
                        leader: { name: 'foo', addr: 'localhost:9002' },
                        segments: 1
                    }
                ],
                messages: {
                    'shopOrder': {
                        name: 'shopOrder',
                        title: 'Shop New Order notification',
                        summary: 'A message containing details of a new order',
                        description: 'More info about how the order notifications are **created** and **used**.',
                        key: { schema: { type: 'string' } },
                        payload: { schema: Product },
                        contentType: 'application/json'
                    }
                },
                bindings: {
                    partitions: 3,
                    retentionBytes: 1500,
                    retentionMs: 30000,
                    segmentBytes: 30000,
                    segmentMs: 1500,
                    valueSchemaValidation: true,
                    keySchemaValidation: false
                }
            },
            {
                name: 'mokapi.shop.userSignedUp',
                summary: 'This channel contains a message per each user who signs up in our application.',
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 0,
                        leader: { name: 'foo', addr: 'localhost:9002' },
                        segments: 1
                    }
                ],
                messages: {
                    'userSignedUp': {
                        name: 'userSignedUp',
                        title: 'title',
                        key: { schema: { type: 'string' } },
                        payload: { schema: UserSignedUp, format: 'application/vnd.oai.openapi+json;version=3.0.0' },
                        contentType: 'application/xml'
                    },
                    'second': {
                        name: 'second',
                        key: { schema: { type: 'string' } },
                        payload: { schema: SecondMessage },
                        contentType: 'application/json'
                    }
                }
            },
            {
                name: 'mokapi.shop.avro',
                description: 'A channel with avro messages',
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 0,
                        leader: { name: 'foo', addr: 'localhost:9002' },
                        segments: 1
                    }
                ],
                messages: {
                    'avro': {
                        name: 'avro',
                        title: 'avro',
                        key: { schema: { type: 'string' }, format: 'application/vnd.apache.avro;version=1.9.0' },
                        payload: {
                            schema: {
                                type: 'record', fields: [
                                    { name: 'Name', type: 'string' },
                                    { name: 'Age', type: 'int' }
                                ]
                            }, format: 'application/vnd.apache.avro;version=1.9.0'
                        },
                        contentType: 'avro/binary'
                    },
                }
            },
            {
                name: 'orders.{region}.{eventType}',
                messages: {
                    'orderMessage': {
                        name: 'OrderEvent',
                        title: 'Order Event Payload',
                        payload: { schema: Order },
                        contentType: 'application/json'
                    }
                },
                parameters: [ 'region', 'eventType' ],
                instances: [
                    {
                        name: 'orders.us.created',
                        parameters: {
                            region: 'us',
                            eventType: 'created'
                        },
                    }
                ],
            },
            {
                name: 'orders.us.created',
                template: 'orders.{region}.{eventType}',
                messages: {
                    'orderMessage': {
                        name: 'OrderEvent',
                        title: 'Order Event Payload',
                        payload: { schema: Order },
                        contentType: 'application/json'
                    }
                },
                parameters: { eventType: 'created', region: 'eu' },
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 0,
                        leader: { name: 'foo', addr: 'localhost:9002' },
                        segments: 1
                    }
                ],
            }
        ],
        groups: [
            {
                name: 'foo',
                generation: 0,
                members: [
                    {
                        name: 'julie',
                        clientId: 'consumer-1',
                        addr: '127.0.0.1:15001',
                        software: 'mokapi 1.0',
                        heartbeat: 1654771269,
                        partitions: { 'mokapi.shop.products': [0, 1], 'mokapi.shop.userSignedUp': [0] }
                    },
                    {
                        name: 'hermann',
                        clientId: 'consumer-2',
                        addr: '127.0.0.1:15002',
                        software: 'mokapi 1.0',
                        heartbeat: 1654872269,
                        partitions: { 'mokapi.shop.products': [2], 'mokapi.shop.userSignedUp': [] }
                    }
                ],
                leader: 'julie',
                state: 'Stable',
                protocol: 'Range',
                topics: ['mokapi.shop.products', 'mokapi.shop.userSignedUp'],
                metrics: {
                    kafka_rebalance_timestamp: metrics.find(m => m.name === `kafka_rebalance_timestamp{service="Kafka World",group="foo"}`)?.value || 0,
                    topics: {
                        'mokapi.shop.products': filterMetrics('kafka_consumer_group_commit', { name: 'group', value: 'foo' }, { name: 'topic', value: 'mokapi.shop.products' }).map(m => {
                            const labels = parseLabels(m)
                            return {
                                partition: labels.partition,
                                kafka_consumer_group_lag: metrics.find(x => x.name === `kafka_consumer_group_lag{service="Kafka World",group="foo",topic="mokapi.shop.products",partition="${labels.partition}"}`)?.value || 0,
                                kafka_consumer_group_commit: m.value
                            }
                        }),
                        'mokapi.shop.userSignedUp': filterMetrics('kafka_consumer_group_commit', { name: 'group', value: 'foo' }, { name: 'topic', value: 'mokapi.shop.userSignedUp' }).map(m => {
                            const labels = parseLabels(m)
                            return {
                                partition: labels.partition,
                                kafka_consumer_group_lag: metrics.find(x => x.name === `kafka_consumer_group_lag{service="Kafka World",group="foo",topic="mokapi.shop.userSignedUp",partition="${labels.partition}"}`)?.value || 0,
                                kafka_consumer_group_commit: m.value
                            }
                        })
                    }
                }
            },
            {
                name: 'bar',
                generation: 0,
                members: [
                    {
                        name: 'george',
                        clientId: 'consumer',
                        addr: '127.0.0.1:15003',
                        software: 'mokapi 1.0',
                        heartbeat: 1654721269,
                        partitions: { 'mokapi.shop.userSignedUp': [0] }
                    }
                ],
                leader: 'george',
                state: 'Stable',
                protocol: 'Range',
                topics: ['mokapi.shop.userSignedUp'],
                metrics: {
                    kafka_rebalance_timestamp: metrics.find(m => m.name === `kafka_rebalance_timestamp{service="Kafka World",group="mokapi.shop.userSignedUp"}`)?.value || 0,
                    topics: {
                        'mokapi.shop.userSignedUp': filterMetrics('kafka_consumer_group_commit', { name: 'group', value: 'bar' }, { name: 'topic', value: 'mokapi.shop.userSignedUp' }).map(m => {
                            const labels = parseLabels(m)
                            return {
                                partition: labels.partition,
                                kafka_consumer_group_lag: metrics.find(x => x.name === `kafka_consumer_group_lag{service="Kafka World",group="foo",topic="mokapi.shop.userSignedUp",partition="${labels.partition}"}`)?.value || 0,
                                kafka_consumer_group_commit: m.value
                            }
                        })
                    }
                }
            },
        ],
        metrics: metrics.filter(x => x.name.includes("kafka")),
        configs: [
            configs['b6fea8ac-56c7-4e73-a9c0-6337640bdca8']
        ],
        clients: [
            {
                clientId: 'consumer-1',
                address: '127.0.0.1:15001',
                software: 'mokapi 1.0',
                groups: [
                    { memberId: 'julie', group: 'foo' }
                ]
            },
            {
                clientId: 'consumer-2',
                address: '127.0.0.1:15002',
                software: 'mokapi 1.0',
                groups: [
                    { memberId: 'hermann', group: 'foo' }
                ]
            },
            {
                clientId: 'consumer',
                address: '127.0.0.1:15003',
                software: 'mokapi 1.0',
                groups: [
                    { memberId: 'george', group: 'bar' }
                ]
            }
        ]
    }
]

export let events = [
    {
        id: '5d549ffe-97cd-477d-8db6-27ed4963c08d',
        traits: {
            namespace: 'kafka',
            name: 'Kafka World',
            type: 'message',
            topic: 'mokapi.shop.products'
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            topic: 'mokapi.shop.products',
            offset: 0,
            key: {
                value: 'GGOEWXXX0827'
            },
            message: {
                value: JSON.stringify(
                    {
                        id: 'GGOEWXXX0827',
                        name: 'Waze Women\'s Short Sleeve Tee',
                        description: 'Made of soft tri-blend jersey fabric, this great t-shirt will help you find your Waze. Made in USA.',
                        features: '<p>Jersey knit</p>\\n<p>37.5% cotton, 50% polyester, 12.5% rayon</p>\\n<p>Made in the USA</p>',
                        price: '18.99',
                        keywords: 'Waze Women\'s Short Sleeve Tee, Waze Short Sleeve Tee, Waze Women\'s Tees, Waze Women\'s tee, waze ladies tees, waze ladies tee, waze short sleeve tees, waze short sleeve tee',
                        url: 'Waze+Womens+Short+Sleeve+Tee',
                        category: 'apparel',
                        subcategory: 'apparel'
                    })
            },
            partition: 0,
            messageId: 'shopOrder',
            headers: {
                foo: { value: 'bar', binary: '' },
                bar: { value: '', binary: 'dmFsdWUgYXMgYmluYXJ5' }
            },
            actions: [
                {
                    duration: 20,
                    tags: {
                        name: "kafka",
                        file: "examples/mokapi/kafka.js",
                        fileKey: "b6fea8ac-56c7-4e73-a9c0-6337640bdca8",
                        event: "kafka"
                    }
                }
            ]
        }
    },
    {
        id: '670ce171-808c-4f8e-a7cd-209b008aada5',
        traits: {
            namespace: 'kafka',
            name: 'Kafka World',
            type: 'message',
            topic: 'mokapi.shop.products'
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            topic: 'mokapi.shop.products',
            offset: 1,
            key: {
                value: 'GGOEWXXX0828'
            },
            message: {
                value: JSON.stringify(
                    {
                        id: 'GGOEWXXX0828',
                        name: 'Waze Men\'s Short Sleeve Tee',
                        description: 'Made of soft tri-blend jersey fabric, this great t-shirt will help you find your Waze. Made in USA.',
                        features: '<p>Jersey knit</p>\\n<p>37.5% cotton, 50% polyester, 12.5% rayon</p>\\n<p>Made in the USA</p>',
                        price: '18.99',
                        keywords: 'Waze Men\'s Short Sleeve Tee, Waze Short Sleeve Tee, Waze Men\'s Tees, Waze Men\'s tee, waze mens tees, waze mens tee, waze short sleeve tees, waze short sleeve tee',
                        url: 'Waze+Mens+Short+Sleeve+Tee',
                        category: 'apparel',
                        subcategory: 'apparel'
                    })
            },
            partition: 1,
            messageId: 'shopOrder',
            deleted: true,
            producerId: 3,
            producerEpoch: 1,
            sequenceNumber: 1
        }
    },
    {
        id: 'cff487df-1fe0-4e1b-b0f6-e599bc61283c',
        traits: {
            namespace: 'kafka',
            name: 'Kafka World',
            type: 'message',
            topic: 'mokapi.shop.avro'
        },
        time: '2025-02-13T09:49:25.482366+01:00',
        data: {
            topic: 'mokapi.shop.avro',
            offset: 1,
            key: {
                binary: 'VEVTVDEgw6Qg8JCAgA=='
            },
            message: {
                value: JSON.stringify(
                    {
                        Name: 'Carol',
                        Age: 29,
                    }),
                binary: base64.encode([0xa, 0x43, 0x61, 0x72, 0x6f, 0x6c, 0x3a].map(x => String.fromCharCode(x)).join(''))
            },
            partition: 0,
            messageId: 'avro'
        }
    },
    {
        id: '84490626-b806-4420-898c-d6b28d1f4ab1',
        traits: {
            namespace: 'kafka',
            name: 'Kafka World',
            type: 'request',
            clientId: 'consumer',
        },
        time: '2025-02-13T09:49:25.482366+01:00',
        data: {
            "api": "Kafka World",
            "header": {
                "requestKey": 10,
                "requestName": "FindCoordinator",
                "version": 2
            },
            "request": {
                "key": "mokapi.shop.products",
                "keyType": 0
            },
            "response": {
                "errorCode": "",
                "errorMessage": "",
                "host": "localhost",
                "port": 9092
            }
        }
    },
    {
        id: '5d549ffe-97cd-477d-8db6-27ed4963c08d',
        traits: {
            namespace: 'kafka',
            name: 'Kafka World',
            type: 'message',
            topic: 'orders.{region}.{eventType}',
            instance: 'orders.us.created'
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            topic: 'orders.us.created',
            offset: 0,
            key: {
                value: 'a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d'
            },
            message: {
                value: JSON.stringify(
                    {
                        orderId: 'a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d',
                        amount: 149.99,
                        timestamp: '2026-10-17T10:15:30Z'
                    })
            },
            partition: 0,
            messageId: 'shopOrder',
        }
    },
]
