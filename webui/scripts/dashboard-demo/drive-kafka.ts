import { Kafka } from 'kafkajs';

const producerClient = new Kafka({
  clientId: 'producer-1',
  brokers: ['localhost:9092'],
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
      key: '7e2b4d0f-1aa1-49e3-9f11-54c266d480c3',
      value: JSON.stringify({
        shipmentId: '8f14e45f-ceea-467f-a8f2-3c6e5b1d2a90',
        orderId: '7e2b4d0f-1aa1-49e3-9f11-54c266d480c3',
        region: 'eu',
        carrier: 'DHL',
        status: 'IN_TRANSIT',
        location: 'Frankfurt Hub',
        timestamp: '2026-10-20T09:15:00Z',
        estimatedDelivery: '2026-10-22'
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