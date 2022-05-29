 import g from 'generator'
import {metrics} from 'metrics';

export let clusters = [
    {
        name:'Kafka World',
        description: g.new({type: 'string', format: '{sentence:10}'}),
        version:g.new({type: 'string', pattern: '[0-9]\\.[0-9]{2}'}),
        contact:{
            name:'mokapi',
            url:'https://www.mokapi.io',
            email:'info@mokapi.io'
        },
        topics:[
            {
                name:'foo',
                description: g.new({type: 'string', format: '{sentence:10}'}),
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
                ]
            },
            {
                name:'bar',
                description: g.new({type: 'string', format: '{sentence:10}'}),
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
                topics: ['foo','bar'],
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
        metrics: metrics.filter(x => x.name.includes("kafka"))
    }
]

export let events = [
     {
         id: "123456",
         traits: {
             namespace: "kafka",
             topic: "foo"
         },
         time: 1651771269,
         data: {
             offset: 0,
             key: "foo",
             message: "{\"id\": 12345}",
         }
     }
 ]
