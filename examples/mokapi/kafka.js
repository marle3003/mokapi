import { metrics } from 'metrics.js'

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
    xml: {
        name: 'user'
    }
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
                url: 'localhost:9092',
                tags: [{name: 'env:test', description: 'This environment is for running internal tests'}],
                description: 'Dashwood contempt on mr unlocked resolved provided of of. Stanhill wondered it it welcomed oh. Hundred no prudent he however smiling at an offence. If earnestly extremity he he propriety something admitting convinced ye.'
            }
        ],
        topics: [
            {
                name: 'mokapi.shop.products',
                description: 'Though literature second anywhere fortnightly am this either so me.',
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
                configs: {
                    name: 'shopOrder',
                    title: 'Shop New Order notification',
                    summary: 'A message containing details of a new order',
                    description: 'More info about how the order notifications are **created** and **used**.',
                    key: { type: 'string' },
                    message: Product,
                    messageType: 'application/json'
                }
            },
            {
                name: 'mokapi.shop.userSignedUp',
                description: 'This channel contains a message per each user who signs up in our application.',
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 0,
                        leader: { name: 'foo', addr: 'localhost:9002' },
                        segments: 1
                    }
                ],
                configs: {
                    name: 'userSignedUp',
                    key: { type: 'string' },
                    message: UserSignedUp,
                    messageType: 'application/xml'
                }
            }
        ],
        groups: [
            {
                name: 'foo',
                members:[
                    {
                        name: 'julie',
                        addr: '127.0.0.1:15001',
                        clientSoftwareName: 'mokapi',
                        clientSoftwareVersion: '1.0',
                        heartbeat: 1654771269,
                        partitions: { 'mokapi.shop.products': [ 0,1 ], 'mokapi.shop.userSignedUp': [ 0 ] }
                    },
                    {
                        name: 'hermann',
                        addr: '127.0.0.1:15002',
                        clientSoftwareName: 'mokapi',
                        clientSoftwareVersion: '1.0',
                        heartbeat: 1654872269,
                        partitions: { 'mokapi.shop.products': [ 2 ], 'mokapi.shop.userSignedUp': [ ] }
                    }
                ],
                coordinator: 'localhost:9092',
                leader: 'julie',
                state: 'Stable',
                protocol: 'Range',
                topics: [ 'mokapi.shop.products', 'mokapi.shop.userSignedUp' ],
            },
            {
                name: 'bar',
                members: [
                    {
                        name: 'george',
                        addr: '127.0.0.1:15003',
                        clientSoftwareName: 'mokapi',
                        clientSoftwareVersion: '1.0',
                        heartbeat: 1654721269,
                        partitions: { 'mokapi.shop.userSignedUp': [ 0 ] }
                    }
                ],
                coordinator: 'localhost:9092',
                leader: 'george',
                state: 'Stable',
                protocol: 'Range',
                topics: [ 'mokapi.shop.userSignedUp' ],
            },
        ],
        metrics: metrics.filter(x => x.name.includes("kafka")),
        configs: [
            configs[ 'b6fea8ac-56c7-4e73-a9c0-6337640bdca8' ]
        ]
    }
]

export let events = [
     {
         id: '123456',
         traits: {
             namespace: 'kafka',
             name: 'Kafka World',
             topic: 'mokapi.shop.products'
         },
         time: '2023-02-13T09:49:25.482366+01:00',
         data: {
             offset: 0,
             key: 'GGOEWXXX0827',
             message: JSON.stringify({
                         id: 'GGOEWXXX0827',
                         name: 'Waze Women\'s Short Sleeve Tee',
                         description: 'Made of soft tri-blend jersey fabric, this great t-shirt will help you find your Waze. Made in USA.',
                         features: '<p>Jersey knit</p>\\n<p>37.5% cotton, 50% polyester, 12.5% rayon</p>\\n<p>Made in the USA</p>',
                         price: '18.99',
                         keywords: 'Waze Women\'s Short Sleeve Tee, Waze Short Sleeve Tee, Waze Women\'s Tees, Waze Women\'s tee, waze ladies tees, waze ladies tee, waze short sleeve tees, waze short sleeve tee',
                         url: 'Waze+Womens+Short+Sleeve+Tee',
                         category: 'apparel',
                         subcategory: 'apparel'
                         }),
             partition: 0,
             headers: {
                 foo: 'bar'
             }
         }
     },
    {
        id: '323456',
        traits: {
            namespace: 'kafka',
            name: 'Kafka World',
            topic: 'mokapi.shop.products'
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            offset: 1,
            key: 'GGOEWXXX0828',
            message: JSON.stringify({
                        id: 'GGOEWXXX0828',
                        name: 'Waze Men\'s Short Sleeve Tee',
                        description: 'Made of soft tri-blend jersey fabric, this great t-shirt will help you find your Waze. Made in USA.',
                        features: '<p>Jersey knit</p>\\n<p>37.5% cotton, 50% polyester, 12.5% rayon</p>\\n<p>Made in the USA</p>',
                        price: '18.99',
                        keywords: 'Waze Men\'s Short Sleeve Tee, Waze Short Sleeve Tee, Waze Men\'s Tees, Waze Men\'s tee, waze mens tees, waze mens tee, waze short sleeve tees, waze short sleeve tee',
                        url: 'Waze+Mens+Short+Sleeve+Tee',
                        category: 'apparel',
                        subcategory: 'apparel'
                      }),
            partition: 1
        }
    }
 ]
