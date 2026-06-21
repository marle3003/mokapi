import { on, every } from 'mokapi';
import kafka from 'mokapi/kafka';

export default async function() {
    let counter = 1;
    every('5m', async function() {
        const result = await kafka.produceAsync({
            topic: 'order-topic',
            messages: [
                { key: 'random-message-' + counter++ }
            ],
        });
        console.log(`offset=${result.messages[0].offset}, key=${result.messages[0].key}, value=${result.messages[0].value}`)
    }, { times: 2, tags: { custom: 'random message generator' } });

    on('kafka', (msg) => {
        msg.headers['correlationId'] = 'abc-123'
    })
}