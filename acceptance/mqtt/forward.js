import { on } from 'mokapi';
import { publish } from 'mokapi/mqtt';

export default function() {
    on('mqtt', (event) => {
        // only forward readings that ask for it, otherwise the handler would react to its own publish
        if (event.topic !== 'sensors/temperature') {
            return;
        }
        let reading;
        try {
            reading = JSON.parse(event.value);
        } catch {
            return;
        }
        if (reading.sensorId !== 's-forward') {
            return;
        }
        publish({
            cluster: 'Sensor Service',
            topic: 'sensors/alerts',
            retain: true,
            value: event.value,
        });
    });
}
