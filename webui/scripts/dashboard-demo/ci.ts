import { collectDashboard } from './collect-dashboard';
import { driveHttp } from './drive-http';
import { driveKafka, closeKafka } from './drive-kafka';
import { driveMqtt } from './drive-mqtt';
import { driveWebsocket } from './drive-websocket'
import { driveMail } from './drive-mail';
import { driveLdap } from './drive-ldap';
import { collectBruno } from './bruno';

async function main() {
    try {
        await driveHttp();
        await driveKafka();
        await driveMqtt();
        await driveWebsocket();
        await driveMail();
        await driveLdap();

        await collectDashboard();
        await collectBruno();

    } catch (err) {
        console.error('❌ Failed to generate demo data')
        console.error(err)
        process.exitCode = 1
    } finally {
        await closeKafka();
    }

    process.exit()
}

await main()