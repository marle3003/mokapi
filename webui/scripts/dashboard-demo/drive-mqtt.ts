import mqtt from 'mqtt';

export async function driveMqtt(): Promise<void> {
    const client = mqtt.connect('mqtt://localhost:1883')

    await new Promise<void>((resolve, reject) => {
        client.on('connect', () => {
            console.log('mqtt: subscribe topic')

            client.subscribe('sensors/12345/data', (err) => {
                if (err) {
                    console.error('mqtt: ', err)
                    reject(err);
                    return;
                }

                console.log('mqtt: publish message')
                client.publish('sensors/12345/data', JSON.stringify({
                    temp: 24,
                    timestamp: '2026-02-13T09:49:25.482366+01:00'
                }));
            });
        });

        client.on('message', (topic, message) => {
            console.log(message.toString());
            client.end();
            resolve();
        });

        client.on('error', (err) => {
            console.error('mqtt: ', err)
            reject(err);
        });
    });
}