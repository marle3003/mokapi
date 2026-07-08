import { metrics } from 'metrics.ts';
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

export let services = [
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
        name: '/chat',
        summary: 'Single WS endpoint carrying all chat messages',
        messages: {
          ChatMessage: {
            name: 'ChatMessage',
            title: 'Chat Message',
            payload: { schema: ChatMessage },
            contentType: 'application/json',
          },
        },
      },
      {
        name: '/chats/{chatId}',
        messages: {
          ChatMessage: {
            name: 'ChatMessage',
            title: 'Chat Message',
            payload: { schema: ChatMessage },
            contentType: 'application/json',
          },
        },
        instances: [
          {
            name: 'chats/1234',
            parameters: {
              chatId: '1234'
            }
          }
        ]
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
              ...(x.title ? { title: x.title } : {}),
              ...(x.summary ? { summary: x.summary } : {}),
              ...(x.description ? { description: x.description } : {}),
              messages: x.messages,
               ...x.instances ? { instances: x.instances } : {},
              metrics: {
                websocket_messages_total: metrics.find(
                  (m) => m.name === `websocket_messages_total{service="${services[0].name}",channel="${x.name}"}`
                )?.value,
                websocket_message_timestamp: metrics.find(
                  (m) => m.name === `websocket_message_timestamp{service="${services[0].name}",channel="${x.name}"}`
                )?.value,
              },
            };
          }),
          clients: services[0].clients,
        };
    }
  });
}

export let events = [
  {
    id: 'a1b2c3d4-0001-4d2d-a1b4-f022aeddeaf8',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chat',
      type: 'message',
      clientId: '67bab4de-e477-4afe-8696-df3102f7a8d8',
    },
    time: '2026-02-13T09:49:25.482366+01:00',
    data: {
      channel: '/chat',
      message: {
        value: '{"userId":"alice","username":"Alice","text":"Hello, world!","timestamp":"2026-02-13T09:49:25.482366+01:00"}',
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      client: { ...clients[0], direction: 'send', server: clients[0].serverAddress }
    },
  },
  {
    id: 'a1b2c3d4-0002-4d2d-a1b4-f022aeddeaf8',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chat',
      type: 'message',
      clientId: 'b4788220-b169-483c-8406-498fbeb482fd',
    },
    time: '2026-02-13T09:49:26.100000+01:00',
    data: {
      channel: '/chat',
      message: {
        value: '{"userId":"bob","username":"Bob","text":"Hi Alice!","timestamp":"2026-02-13T09:49:26.100000+01:00"}',
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      client: { ...clients[1], direction: 'send', server: clients[1].serverAddress }
    }
  },
  {
    id: '5cc3940e-8b3f-4c65-9f85-9fd6f7e51e5e',
    traits: {
      namespace: 'websocket',
      name: 'WebSocket Chat API',
      channel: '/chats/{chatId}',
      type: 'message',
      clientId: '5cc3940e-8b3f-4c65-9f85-9fd6f7e51e5e',
    },
    time: '2026-02-13T09:49:26.100000+01:00',
    data: {
      channel: '/chats/1234',
      message: {
        value: '{"userId":"carol","username":"Carol","text":"Hi Alice!","timestamp":"2026-02-13T09:49:26.100000+01:00"}',
      },
      messageId: 'ChatMessage',
      api: 'WebSocket Chat API',
      client: { ...clients[1], direction: 'send', server: clients[1].serverAddress },
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
    }
  },
];