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
const producer = producerClient.producer();

export async function driveKafka() {
  await consumer.connect();
  await producer.connect();

  await consumer.subscribe({ topic: 'order-topic', fromBeginning: true });

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const value = JSON.parse(message.value!.toString());
      console.log('Received command:', value);
    }});

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

    console.log('Published order:', order);
};

export async function closeKafka() {
    await consumer.stop();
    await consumer.disconnect();
    await producer.disconnect();
}