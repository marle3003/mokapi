import { collectDashboard } from './collect-dashboard.ts';
import { closeKafka } from './drive-kafka.ts';
import { driveAll } from './drive-all.ts';
import { startMokapi, stopMokapi } from './mokapi.ts';
import { collectBruno } from './bruno.ts';

async function main() {
    try {
        console.log('🚀 Starting Mokapi...');
        // at least 5 s to get a useful memory usage metric
        await startMokapi({ warmUpMs: 5500 });

        await driveAll();

        await collectDashboard();
        await collectBruno();

    } catch (err) {
        console.error('❌ Failed to generate demo data')
        console.error(err)
        process.exitCode = 1
    } finally {
        await closeKafka();
        stopMokapi();
    }
}

main()