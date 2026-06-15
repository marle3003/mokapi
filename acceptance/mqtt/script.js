import { publishAsync } from 'mokapi/mqtt';

export default async function() {
    publishAsync({
        cluster: 'Sensor Service',
        topic: 'sensors/temperature',
        retain: true,
        value: JSON.stringify({
            sensorId: 's123',
            value: 12.3
        }),
    })
}