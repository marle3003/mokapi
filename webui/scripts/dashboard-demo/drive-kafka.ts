import { Kafka } from 'kafkajs';

const producerClient = new Kafka({
  clientId: 'producer-1',
  brokers: ['localhost:9092']
});

const consumerClient = new Kafka({
  clientId: 'consumer-1',
  brokers: ['localhost:9092']
});

const consumer = consumerClient.consumer({ groupId: 'order-status-group-100' });
const producer = producerClient.producer({
  idempotent: true,
  allowAutoTopicCreation: true
});

export async function driveKafka() {
  await consumer.connect();
  await producer.connect();

  await consumer.subscribe({ topic: 'order-topic', fromBeginning: true });

  await consumer.run({
    eachMessage: async ({ message }) => {
      const value = JSON.parse(message.value!.toString());
      console.log('Received command:', value);
    }
  });

  const order = {
    orderId: 'a914817b-c5f0-433e-8280-1cd2fe44234e',
    productId: '2adb46de-1c96-4290-a215-9701b0a7900c',
    productName: 'Wireless Ergonomic Mouse (Black)',
    quantity: 65,
    customerEmail: 'johnnydibbert@feil.org',
    status: 'SHIPPED'
  };

  await producer.send({
    topic: 'order-topic',
    messages: [{ key: order.orderId, value: JSON.stringify(order) }]
  });

  await producer.send({
    topic: 'shipments.eu.updates',
    messages: [{
      key: '8f14e45f-ceea-467f-a8f2-3c6e5b1d2a90',
      value: JSON.stringify({
        shipmentId: '8f14e45f-ceea-467f-a8f2-3c6e5b1d2a90',
        orderId: 'c9bf9e57-1685-4c89-bafb-ff5af830be8a',
        region: 'eu',
        carrier: 'DHL',
        status: 'IN_TRANSIT',
        location: 'Frankfurt Hub',
        timestamp: '2026-10-20T09:15:00Z',
        estimatedDelivery: '2026-10-22'
      })
    }],
  });

  await producer.send({
    topic: 'shipments.us.updates',
    messages: [{
      key: 'a3bb189e-8bf9-4888-9912-ace4e6543002',
      value: JSON.stringify({
        shipmentId: '1b4e28ba-2fa1-41d2-883f-0016d3cca427',
        orderId: 'a3bb189e-8bf9-4888-9912-ace4e6543002',
        region: 'us',
        carrier: 'UPS',
        status: 'DELIVERED',
        location: 'Austin, TX',
        timestamp: '2026-10-20T16:42:00Z'
      })
    }],
  });

  console.log('Published order:', order);
};

export async function closeKafka() {
  await consumer.stop();
  await consumer.disconnect();
  await producer.disconnect();
}