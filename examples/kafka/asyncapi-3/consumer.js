import { Kafka } from 'kafkajs';

const consumerClient = new Kafka({
    clientId: 'consumer-1',
    brokers: ['localhost:9092']
});

const consumer = consumerClient.consumer({ groupId: 'group-1' });
await consumer.connect();
await consumer.subscribe({ topic: 'users.signedup', fromBeginning: true });

await consumer.run({
    eachMessage: async ({ message }) => {
        const value = JSON.parse(message.value.toString());
        console.log('Received command:', value);
    }
});