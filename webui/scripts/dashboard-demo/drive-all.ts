import { driveHttp } from './drive-http.ts';
import { driveKafka } from './drive-kafka.ts';
import { driveMqtt } from './drive-mqtt.ts';
import { driveWebsocket } from './drive-websocket.ts'
import { driveMail } from './drive-mail.ts';
import { driveLdap } from './drive-ldap.ts';

// The Kafka client stays connected, the caller closes it with closeKafka() when it is done.
export async function driveAll() {
    await driveHttp();
    await driveKafka();
    await driveMqtt();
    await driveWebsocket();
    await driveMail();
    await driveLdap();
}
