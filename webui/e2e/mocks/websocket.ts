import { sum, max, metrics } from './metrics';
import { on } from 'mokapi';

const ChatMessage = {
  type: 'object',
  properties: {
    userId: { type: 'string' },
    username: { type: 'string' },
    text: { type: 'string' },
    timestamp: { type: 'string', format: 'date-time' },
  },
};

const clients = [
  { id: '67bab4de-e477-4afe-8696-df3102f7a8d8', address: '127.0.0.1:53211', serverAddress: 'localhost:8080' },
  { id: 'b4788220-b169-483c-8406-498fbeb482fd', address: '127.0.0.1:53298', serverAddress: 'localhost:8080' },
  { id: '5cc3940e-8b3f-4c65-9f85-9fd6f7e51e5e', address: '127.0.0.1:53200', serverAddress: 'localhost:8080' },
]

const chatTag = { name: 'chat', description: 'Chat messages between users' }
const broadcastTag = { name: 'broadcast', description: 'Messages for every connected client' }

export const services = [
  {
    name: 'WebSocket Chat API',
    description:
      'A simple chat app mocked over WebSocket. Clients open a single connection to /chat and exchange text frames in real time.',
    version: '1.0.0',
    contact: {
      name: 'mokapi',
      url: 'https://www.mokapi.io',
      email: 'info@mokapi.io',
    },
    servers: [
      {
        name: 'Chat Server',
        host: 'localhost:8080',
        path: '/chat',
        tags: [
          {
            name: 'env:test',
            description: 'This environment is for running internal tests',
          },
        ],
        description: 'Development server',
      },
    ],
    channels: [
      {
        name: '/chats/{chatId}',
        tags: [chatTag],
        messages: {
          ChatMessage: {
            name: 'ChatMessage',
            title: 'Chat Message',
            payload: { schema: ChatMessage },
            contentType: 'application/json',
          },
          Foo: {
            name: 'ChatMessage',
            title: 'Chat Message',
            payload: { schema: { type: 'object', required: ['foo', 'bar'], properties: { foo: { type: 'string' }, bar: { type: 'number'} } } },
            contentType: 'application/json',
          }
        },
        parameters: ['chatId'],
        instances: [
          {
            name: '/chats/1234',
            parameters: {
              chatId: '1234'
            },
          },
          {
            name: '/chats/5678',
            parameters: {
              chatId: '5678'
            },
          }
        ],
      },
      {
        name: '/chat',
        tags: [chatTag, broadcastTag],
        summary: 'Single WS endpoint carrying all chat messages',
        messages: {
          ChatMessage: {
            name: 'ChatMessage',
            title: 'Chat Message',
            payload: { schema: ChatMessage },
            contentType: 'application/json',
          }
        },
      },
    ],
    metrics: metrics.filter((x) => x.name.includes('websocket')),
    clients: clients
  },
];

export default async function () {
  on('http', function (request, response) {
    switch (request.operationId) {
      case 'serviceWebsocket':
        response.data = {
          name: services[0].name,
          description: services[0].description,
          version: services[0].version,
          contact: services[0].contact,
          servers: services[0].servers,
          channels: services[0].channels.map((x) => {
            return {
              name: x.name,
              ...('title' in x ? { title: x.title } : {}),
              ...('summary' in x ? { summary: x.summary } : {}),
              ...('description' in x ? { description: x.description } : {}),
              ...('tags' in x ? { tags: x.tags } : {}),
              messages: x.messages,
              ...('parameters' in x ? { parameters: x.parameters} : {}), 
              ...('instances' in x ? { instances: x.instances.map((i) => (
                {
                  ...i,
                  metrics: {
                    messages_total: sum('websocket_messages_total', { name: 'service', value: services[0].name }, { name: 'channel', value: x.name }, { name: 'instance', value: i.name }),
                    message_timestamp: max('websocket_message_timestamp', { name: 'service', value: services[0].name }, { name: 'channel', value: x.name }, { name: 'instance', value: i.name })
                  },
                }
              )) } : {}),
              metrics: {
                messages_total: sum('websocket_messages_total', { name: 'service', value: services[0].name }, { name: 'channel', value: x.name }),
                message_timestamp: max('websocket_message_timestamp', { name: 'service', value: services[0].name }, { name: 'channel', value: x.name })
              },
            };
          }),
          clients: services[0].clients,
        };
    }
  });
}

const handlerActions = [
  {
    duration: 20,
    tags: {
      name: 'websocket',
      file: 'examples/mokapi/websocket.js',
      fileKey: 'b6fea8ac-56c7-4e73-a9c0-6337640bdca8',
      event: 'websocket',
    },
  },
]

// cron.js of the config mock (jobs.ts), so the Script link of a message resolves
const scriptConfigId = '6f5cab53-9f51-4bbe-bb32-fe41ee71d542'

interface MessageSeed {
  id: string
  time: string
  channel: string
  instance?: string
  client: (typeof clients)[number]
  direction: 'send' | 'receive'
  username: string
  text: string
  withActions?: boolean
  script?: string
}

// send: the client sent the message to the server, receive: the server sent it to the client
function messageEvent(seed: MessageSeed) {
  return {
    id: seed.id,
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: seed.channel,
      ...(seed.instance ? { instance: seed.instance } : {}),
      type: 'message',
      clientId: seed.client.id,
    },
    time: seed.time,
    data: {
      channel: seed.instance ?? seed.channel,
      message: {
        value: JSON.stringify({ userId: seed.username.toLowerCase(), username: seed.username, text: seed.text, timestamp: seed.time }),
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      direction: seed.direction,
      client: { ...seed.client, server: seed.client.serverAddress },
      ...(seed.withActions ? { actions: handlerActions } : {}),
      ...(seed.script ? { script: seed.script } : {}),
    },
  }
}

function connectionEvent(seed: { id: string, time: string, type: 'connect' | 'close', channel: string, client: (typeof clients)[number], reason?: string, closedBy?: string, withActions?: boolean }) {
  return {
    id: seed.id,
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: seed.channel,
      type: 'connection',
      clientId: seed.client.id,
    },
    time: seed.time,
    data: {
      type: seed.type,
      channel: seed.channel,
      ...(seed.reason ? { reason: seed.reason, closedBy: seed.closedBy } : {}),
      client: {
        id: seed.client.id,
        query: {},
        header: {},
        address: seed.client.address,
        server: seed.client.serverAddress,
      },
      api: 'WebSocket Chat API',
      ...(seed.withActions ? { actions: handlerActions } : {}),
    },
  }
}

export const events = [
  {
    id: 'a1b2c3d4-0001-4d2d-a1b4-f022aeddeaf8',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chat',
      type: 'message',
      clientId: clients[0].id,
    },
    time: '2026-02-13T09:49:25.482366+01:00',
    data: {
      channel: '/chat',
      message: {
        value: '{"userId":"alice","username":"Alice","text":"Hello, world!","timestamp":"2026-02-13T09:49:25.482366+01:00"}',
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      direction: 'send',
      client: { ...clients[0], server: clients[0].serverAddress }
    },
  },
  {
    id: '4d955c5e-030e-4b29-a762-c3be07d9b7f0',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chat',
      type: 'message',
      clientId: clients[1].id,
    },
    time: '2026-02-13T09:49:26.100000+01:00',
    data: {
      channel: '/chat',
      message: {
        value: '{"userId":"bob","username":"Bob","text":"Hi Alice!","timestamp":"2026-02-13T09:49:26.100000+01:00"}',
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      direction: 'send',
      client: { ...clients[1], server: clients[1].serverAddress }
    }
  },
  {
    id: '5cc3940e-8b3f-4c65-9f85-9fd6f7e51e5e',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chats/{chatId}',
      instance: '/chats/1234',
      type: 'message',
      clientId: clients[1].id,
    },
    time: '2026-02-13T09:49:26.100000+01:00',
    data: {
      channel: '/chats/1234',
      message: {
        value: '{"userId":"carol","username":"Carol","text":"Hi Alice!","timestamp":"2026-02-13T09:49:26.100000+01:00"}',
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      direction: 'send',
      client: { ...clients[1], server: clients[1].serverAddress },
      actions: [
        {
            duration: 20,
            tags: {
                name: "websocket",
                file: "examples/mokapi/websocket.js",
                fileKey: "b6fea8ac-56c7-4e73-a9c0-6337640bdca8",
                event: "websocket"
            }
        }
      ]
    },
  },
  {
    id: '0ea1241e-2468-4f58-9da0-4d0c08bdf6fb',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chats/{chatId}',
      instance: '/chats/1234',
      type: 'message',
      clientId: clients[1].id,
    },
    time: '2026-02-13T09:50:26.100000+01:00',
    data: {
      channel: '/chats/1234',
      message: {
        value: '{"bar": "not a number"}',
      },
      api: 'WebSocket Chat API',
      direction: 'send',
      client: { ...clients[1], server: clients[1].serverAddress },
      validationErrors: {
        ChatMessage: 'Validation error count 1:\n\t- #/type: invalid type, expected object but got string',
        Foo: 'Validation error count 2:\n\t- #/bar/type: invalid type, expected number but got string\n\t- #/required: required properties are missing: foo'
      },
      actions: [
        {
            duration: 20,
            tags: {
                name: "websocket",
                file: "examples/mokapi/websocket.js",
                fileKey: "b6fea8ac-56c7-4e73-a9c0-6337640bdca8",
                event: "websocket"
            }
        }
      ]
    },
  },
  {
    id: '9ab700fa-55b6-4c32-9c53-8cd3926dacc4',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chats/general',
      type: 'connection',
      clientId: clients[0].id,
    },
    time: '2026-02-13T09:50:26.100000+01:00',
    data: {
      type: 'connect',
      channel: '/chats/general',
      client: {
        id: clients[0].id,
        query: {},
        header: {},
        address: clients[0].address,
        server: clients[0].serverAddress
      },
      api: 'WebSocket Chat API',
    },
  },

  // 12 messages in total, 6 per client: /chat 5, /chats/1234 4, /chats/5678 3 (metrics.ts has the same totals)
  messageEvent({ id: 'a1b2c3d4-0005-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:51:00.000000+01:00', channel: '/chat', client: clients[0], direction: 'receive', username: 'Server', text: 'Welcome to the chat', withActions: true }),
  messageEvent({ id: 'a1b2c3d4-0006-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:51:10.000000+01:00', channel: '/chat', client: clients[1], direction: 'send', username: 'Bob', text: 'Anyone here?' }),
  messageEvent({ id: 'a1b2c3d4-0007-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:51:20.000000+01:00', channel: '/chats/{chatId}', instance: '/chats/5678', client: clients[0], direction: 'send', username: 'Alice', text: 'Lunch at noon?' }),
  messageEvent({ id: 'a1b2c3d4-0008-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:51:30.000000+01:00', channel: '/chats/{chatId}', instance: '/chats/5678', client: clients[1], direction: 'receive', username: 'Server', text: 'Alice joined the room', withActions: true }),
  messageEvent({ id: 'a1b2c3d4-0009-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:51:40.000000+01:00', channel: '/chats/{chatId}', instance: '/chats/1234', client: clients[0], direction: 'receive', username: 'Server', text: 'Scheduled announcement', script: scriptConfigId }),
  messageEvent({ id: 'a1b2c3d4-0010-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:51:50.000000+01:00', channel: '/chats/{chatId}', instance: '/chats/5678', client: clients[0], direction: 'send', username: 'Alice', text: 'See you there' }),
  messageEvent({ id: 'a1b2c3d4-0011-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:52:00.000000+01:00', channel: '/chat', client: clients[0], direction: 'send', username: 'Alice', text: 'Good night' }),
  messageEvent({ id: 'a1b2c3d4-0012-4d2d-a1b4-f022aeddeaf8', time: '2026-02-13T09:52:10.000000+01:00', channel: '/chats/{chatId}', instance: '/chats/1234', client: clients[1], direction: 'send', username: 'Bob', text: 'Bye' }),
  connectionEvent({ id: '9ab700fa-0002-4c32-9c53-8cd3926dacc4', time: '2026-02-13T09:52:20.000000+01:00', type: 'connect', channel: '/chats/1234', client: clients[1] }),
  connectionEvent({ id: '9ab700fa-0003-4c32-9c53-8cd3926dacc4', time: '2026-02-13T09:52:30.000000+01:00', type: 'close', channel: '/chats/1234', client: clients[1], reason: 'connection closed', closedBy: 'client', withActions: true }),
];
