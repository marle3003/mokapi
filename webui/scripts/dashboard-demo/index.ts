import { collectDashboard } from './collect-dashboard.ts';
import { driveHttp } from './drive-http.ts';
import { driveKafka, closeKafka } from './drive-kafka.ts';
import { startMokapi, stopMokapi } from './mokapi.ts';
import { driveMail } from './drive-mail.ts';
import { driveLdap } from './drive-ldap.ts';

async function main() {
    try {
        console.log('🚀 Starting Mokapi...');
        await startMokapi();

        await driveHttp();
        await driveKafka();
        await driveMail();
        await driveLdap();

        await collectDashboard();

    } catch (err) {
        console.error('❌ Failed to generate demo data')
        console.error(err)
    } finally {
        await closeKafka();
        stopMokapi();
    }
}

main()