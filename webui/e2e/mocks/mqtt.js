import { metrics } from 'metrics.js';
import { on } from 'mokapi';

const Measurement = {
  type: 'object',
  properties: {
    sensorId: { type: 'string' },
    temperature: { type: 'number', format: 'float' },
    unit: { type: 'string', enum: ['celsius', 'fahrenheit'] },
    timestamp: { type: 'string', format: 'date-time' },
  },
};

export let clusters = [
  {
    name: 'MQTT Temperature Sensor API',
    description:
      'API for an MQTT-based temperature sensor. The sensor publishes measurement data and receives configuration commands.',
    version: '1.0.0',
    contact: {
      name: 'mokapi',
      url: 'https://www.mokapi.io',
      email: 'info@mokapi.io',
    },
    servers: [
      {
        name: 'Broker',
        host: 'localhost:1883',
        tags: [
          {
            name: 'env:test',
            description: 'This environment is for running internal tests',
          },
        ],
        description: 'Development server',
      },
    ],
    topics: [
      {
        name: 'home/livingroom/temperature',
        description: 'Channel for messages FROM the sensor (Publish)',
        messages: {
          TemperatureMeasurement: {
            name: 'TemperatureMeasurement',
            title: 'Temperature Measurement',
            payload: { schema: Measurement },
            contentType: 'application/json',
          },
        },
        bindings: {
          qos: 1,
          retain: false,
        },
      },
      {
        name: 'sensors/{sensorId}/data',
        messages: {
          TemperaturePayload: {
            name: 'TemperaturePayload',
            title: 'Temperature Payload',
            payload: { 
              schema: {
                type: 'object',
                properties: {
                  temp: {
                    type: 'number'
                  },
                  timestamp: {
                    type: 'string',
                    format: 'date-time'
                  }
                }
              }
            },
            contentType: 'application/json',
          },
        },
        instances: [
          {
            name: 'sensors/123/data',
            parameters: {
              sensorId: '123'
            }
          }
        ]
      },
    ],
    metrics: metrics.filter((x) => x.name.includes('mqtt')),
    clients: [
      { clientId: 'mqtt-client-1', address: '127.0.0.1:83374', brokerAddress: 'localhost:1883', protocolVersion: 4 },
      { clientId: 'mqtt-client-2', address: '127.0.0.1:83374', brokerAddress: 'localhost:1883', protocolVersion: 5 }
    ]
  },
];

export default async function () {
  on('http', function (request, response) {
    switch (request.operationId) {
      case 'serviceMqtt':
        response.data = clusters[0];
        return true;
    }
  });
}

export let events = [
  {
    id: '9f8bd24d-1480-4d2d-a1b4-f022aeddeaf8',
    traits: {
      namespace: 'mqtt',
      name: 'MQTT Temperature Sensor API',
      type: 'message',
      topic: 'home/livingroom/temperature',
      clientId: 'mqtt-client-1',
    },
    time: '2026-02-13T09:49:25.482366+01:00',
    data: {
      topic: 'home/livingroom/temperature',
      messageId: 'TemperatureMeasurement',
      message: {
        value: JSON.stringify({
          sensorId: '12345',
          temperature: 30.0,
          unit: 'celsius',
          timestamp: '2026-02-13T09:49:25.482366+01:00'
        }),
      },
      clientId: 'mqtt-client-1',
    },
  },
  {
    id: '0d39cc83-2eda-4b36-9645-107186ae401d',
    traits: {
      namespace: 'mqtt',
      name: 'MQTT Temperature Sensor API',
      type: 'message',
      topic: 'sensors/{sensorId}/data',
      sensorId: '123',
      clientId: 'mqtt-client-2',
    },
    time: '2026-02-14T09:49:25.482366+01:00',
    data: {
      topic: 'sensors/123/data',
      messageId: 'TemperaturePayload',
      message: {
        value: JSON.stringify({
          temp: 33,
          timestamp: '2026-02-14T09:49:25.482366+01:00'
        }),
      },
      clientId: 'mqtt-client-2',
    },
  },
  {
    id: '93834ead-6c61-47eb-8068-c269fb3f606b',
    traits: {
      namespace: 'mqtt',
      name: 'MQTT Temperature Sensor API',
      type: 'request',
      clientId: 'mqtt-client-2',
    },
    time: '2026-02-14T09:49:25.482366+01:00',
    data: {
      type: 1,
      request: {
        version: 5,
        cleanSession: true,
        keepAlive: 60,
        message: {
          qos: 1,
          retain: true,
          topic: 'sensors/123/data',
          message: '{"temp":33,"timestamp":"2026-02-14T09:49:25.482366+01:00}"}'
        }
      },
      response: {
        sessionPresent: false,
        reasonCode: { reason: 'Success', code: 0 }
      }
    },
  },
  {
    id: '8897df7c-c3bb-4e57-8bbe-ef5a1f71d978',
    traits: {
      namespace: 'mqtt',
      name: 'MQTT Temperature Sensor API',
      type: 'request',
      clientId: 'mqtt-client-2',
    },
    time: '2026-02-14T09:49:25.482366+01:00',
    data: {
      type: 8,
      request: {
        messageId: 1234,
        topics: [ { name: 'sensors/123/data', qos: 1 } ]
      },
      response: {
        reasonCodes: [ 1 ]
      }
    },
  },
  {
    id: '469a585e-9254-4c05-bbb6-3369fde147a9',
    traits: {
      namespace: 'mqtt',
      name: 'MQTT Temperature Sensor API',
      type: 'request',
      clientId: 'mqtt-client-2',
    },
    time: '2026-02-14T09:49:25.482366+01:00',
    data: {
      type: 14,
      request: {
        reason: 0
      }
    },
  }
];
