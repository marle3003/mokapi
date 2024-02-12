import {fake} from 'mokapi/faker'
import {metrics} from 'metrics.js'

const Product = {
    type: "object",
    properties: {
        id: {
            type: "string",
        },
        name: {
            type: "string"
        },
        description: {
            type: "string"
        },
        features: {
            type: "string"
        },
        price: {
            type: "number"
        },
        keywords: {
            type: "string"
        },
        url: {
            type: "string"
        },
        category: {
            type: "string"
        },
        subcategory: {
            type: "string"
        }
    }
}

export const configs = {
    'b6fea8ac-56c7-4e73-a9c0-6337640bdca8': {
        id: 'b6fea8ac-56c7-4e73-a9c0-6337640bdca8',
        url: 'file://kafka.json',
        provider: 'file',
        time: '2023-02-15T08:49:25.482366+01:00',
        data: 'http://localhost:8090/api/services/kafka/Kafka%20World'
    }
}

export let clusters = [
    {
        name:'Kafka World',
        description: fake({type: 'string', format: '{sentence:10}'}),
        version:fake({type: 'string', pattern: '[0-9]\\.[0-9]{2}'}),
        contact:{
            name:'mokapi',
            url:'https://www.mokapi.io',
            email:'info@mokapi.io'
        },
        servers: [
            {
                name: "Broker",
                url: "localhost:9092",
                description: "kafka broker"
            }
        ],
        topics:[
            {
                name:'mokapi.shop.products',
                description: fake({type: 'string', format: '{sentence:10}'}),
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 4,
                        leader: {name: 'foo', addr: 'localhost:9002'},
                        segments: 1
                    },
                    {
                        id: 1,
                        startOffset: 0,
                        offset: 3,
                        leader: {'name': 'foo', 'addr': 'localhost:9002'},
                        segments: 1
                    },
                    {
                        id: 2,
                        startOffset: 0,
                        offset: 3,
                        leader: {name: 'foo', addr: 'localhost:9002'},
                        segments: 1
                    }
                ],
                configs: {
                    name: "Products",
                    title: "Products",
                    summary: "Summary",
                    description: "description",
                    key: {type: "string"},
                    message: Product,
                    messageType: "application/json"
                }
            },
            {
                name:'bar',
                description: fake({type: 'string', format: '{sentence:10}'}),
                partitions: [
                    {
                        id: 0,
                        startOffset: 0,
                        offset: 0,
                        leader: {name: 'foo', addr: 'localhost:9002'},
                        segments: 1
                    }
                ]
            }
        ],
        groups:[
            {
                name:'foo',
                members:[
                    {
                        name: 'julie',
                        addr: '127.0.0.1: 15001',
                        clientSoftwareName: 'mokapi',
                        clientSoftwareVersion: '1.0',
                        heartbeat: 1654771269,
                        partitions:[1,2]
                    },
                    {
                        name:'hermann',
                        addr: '127.0.0.1: 15002',
                        clientSoftwareName: 'mokapi',
                        clientSoftwareVersion: '1.0',
                        heartbeat: 1654872269,
                        partitions:[3]
                    }
                ],
                coordinator:'localhost:9092',
                leader:'julie',
                state:'Stable',
                protocol:'Range',
                topics: ['mokapi.shop.products','bar'],
            },
            {
                name:'bar',
                members:[
                    {
                        name:'george',
                        addr: '127.0.0.1: 15003',
                        clientSoftwareName: 'mokapi',
                        clientSoftwareVersion: '1.0',
                        heartbeat: 1654721269,
                        partitions:[1]
                    }
                ],
                coordinator:'localhost:9092',
                leader:'george',
                state:'Stable',
                protocol:'Range',
                topics: ['bar'],
            },
        ],
        metrics: metrics.filter(x => x.name.includes("kafka")),
        configs: [
            configs['b6fea8ac-56c7-4e73-a9c0-6337640bdca8']
        ]
    }
]

export let events = [
     {
         id: "123456",
         traits: {
             namespace: "kafka",
             name: "Kafka World",
             topic: "mokapi.shop.products"
         },
         time: '2023-02-13T09:49:25.482366+01:00',
         data: {
             offset: 0,
             key: "GGOEWXXX0827",
             message: `{
                         "id": "GGOEWXXX0827",
                         "name": "Waze Women's Short Sleeve Tee",
                         "description": "Made of soft tri-blend jersey fabric, this great t-shirt will help you find your Waze. Made in USA.",
                         "features": "<p>Jersey knit</p>\\n<p>37.5% cotton, 50% polyester, 12.5% rayon</p>\\n<p>Made in the USA</p>",
                         "price": "18.99",
                         "keywords": "Waze Women's Short Sleeve Tee, Waze Short Sleeve Tee, Waze Women's Tees, Waze Women's tee, waze ladies tees, waze ladies tee, waze short sleeve tees, waze short sleeve tee",
                         "url": "Waze+Womens+Short+Sleeve+Tee",
                         "category": "apparel",
                         "subcategory": "apparel"
                         }`,
             partition: 0,
             headers: {
                 foo: "bar"
             }
         }
     },
    {
        id: "323456",
        traits: {
            namespace: "kafka",
            name: "Kafka World",
            topic: "mokapi.shop.products"
        },
        time: '2023-02-13T09:49:25.482366+01:00',
        data: {
            offset: 1,
            key: "GGOEWXXX0828",
            message: `{
                        "id": "GGOEWXXX0828",
                        "name": "Waze Men's Short Sleeve Tee",
                        "description": "Made of soft tri-blend jersey fabric, this great t-shirt will help you find your Waze. Made in USA.",
                        "features": "<p>Jersey knit</p>\\n<p>37.5% cotton, 50% polyester, 12.5% rayon</p>\\n<p>Made in the USA</p>",
                        "price": "18.99",
                        "keywords": "Waze Men's Short Sleeve Tee, Waze Short Sleeve Tee, Waze Men's Tees, Waze Men's tee, waze mens tees, waze mens tee, waze short sleeve tees, waze short sleeve tee",
                        "url": "Waze+Mens+Short+Sleeve+Tee",
                        "category": "apparel",
                        "subcategory": "apparel"
                      }`,
            partition: 1
        }
    }
 ]
