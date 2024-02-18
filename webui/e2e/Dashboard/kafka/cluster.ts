import { formatTimestamp } from '../../helpers/format'

export const cluster = {
    name: 'Kafka World',
    version: '4.01',
    contact: {
        name: 'mokapi',
    },
    description: 'To ours significant why upon tomorrow her faithful many motionless.',
    lastMessage: formatTimestamp(1652135690),
    messages: '11',
    brokers: [{
        name: 'Broker',
        url: 'localhost:9092',
        description: 'kafka broker'
    }],
    topics: [
        {
            name: 'mokapi.shop.products',
            description: 'Though literature second anywhere fortnightly am this either so me.',
            lastMessage: formatTimestamp(1652135690),
            messages: '10',
            partitions: [
                {
                    id: '0',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '4',
                    segments: '1'
                },
                {
                    id: '1',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '3',
                    segments: '1'
                },
                {
                    id: '2',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '3',
                    segments: '1'
                }
            ],
            configs: {
                title: 'Shop New Order notification',
                name: 'shopOrder',
                summary: 'A message containing details of a new order',
                description: 'More info about how the order notifications are created and used.',
                contentType: 'application/json',
                message: {
                    lines: '32 lines',
                    size: '281 B'
                }
            }
        },
        {
            name: 'mokapi.shop.userSignedUp',
            description: 'This channel contains a message per each user who signs up in our application.',
            lastMessage: formatTimestamp(1652035690),
            messages: '1',
            partitions: [
                {
                    id: '0',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '0',
                    segments: '1'
                },
            ],
            configs: {
                name: 'userSignedUp',
                contentType: 'application/xml',
                message: {
                    lines: '18 lines',
                    size: '262 B'
                }
            }
        }
    ],
    groups: [
        {
            name: 'foo',
            state: 'Stable',
            protocol: 'Range',
            coordinator: 'localhost:9092',
            leader: 'julie',
            members: [
                {
                    name: 'julie',
                    address: '127.0.0.1:15001',
                    clientSoftware: 'mokapi 1.0',
                    lastHeartbeat: formatTimestamp(1654771269),
                    partitions: [1,2]
                },
                {
                    name: 'herman',
                    address: '127.0.0.1:15002',
                    clientSoftware: 'mokapi 1.0',
                    lastHeartbeat: formatTimestamp(1654872269),
                    partitions: [3]
                }
            ]
        },
        {
            name: 'bar',
            state: 'Stable',
            protocol: 'Range',
            coordinator: 'localhost:9092',
            leader: 'george',
            members: [
                {
                    name: 'george',
                    address: '127.0.0.1:15003',
                    clientSoftware: 'mokapi 1.0',
                    lastHeartbeat: formatTimestamp(1654721269),
                    partitions: [1]
                },
            ]
        }
    ]
}