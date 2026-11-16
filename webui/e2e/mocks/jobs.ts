const cronTags = {
    name: "cron",
    file: "cron.js",
    fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
}
const everyTags = {
    name: "every",
    file: "every.js",
    fileKey: "0c3a5d52-7d6e-4b0f-8a3e-4c7a6d9b1e21",
}
// A job with its own tags: the name comes from the tags, not from the script file
const cleanupTags = {
    name: "cleanup",
    file: "every.js",
    fileKey: "0c3a5d52-7d6e-4b0f-8a3e-4c7a6d9b1e21",
    team: "platform",
}

function jobRun(id: string, name: string, tags: Record<string, string>, time: string, data: Record<string, unknown>) {
    return {
        id,
        traits: { namespace: "job", name },
        time,
        data: { duration: 1500, ...data, tags },
    }
}

export const jobs = [
    {
        id: "8234",
        traits: {
            namespace: "job",
            name: "cron",
        },
        time: new Date(Date.now()).toISOString(),
        data: {
            schedule: '3s',
            maxRuns: -1,
            runs: 1,
            nextRun: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
            duration: 2300,
            tags: {
                name: "cron",
                file: "cron.js",
                fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
            }
        }
    },
    {
        id: "8134",
        traits: {
            namespace: "job",
            name: "cron",
        },
        time: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
        data: {
            schedule: '3s',
            maxRuns: 10,
            runs: 1,
            nextRun: new Date(Date.now()).toISOString(),
            duration: 2300,
            logs: [
                {
                    level: 'error',
                    message: 'an error message, see https://mokapi.io'
                }
            ],
            error: {
                message: 'There was an error'
            },
            tags: {
                name: "cron",
                file: "cron.js",
                fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
            }
        },

    },
    {
        id: "81344",
        traits: {
            namespace: "job",
            name: "cron",
        },
        time: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
        data: {
            schedule: '3s',
            runs: 1,
            nextRun: new Date(Date.now()).toISOString(),
            duration: 2300,
            logs: [
                {
                    level: 'warn',
                    message: 'a warning message, see https://mokapi.io'
                }
            ],
            tags: {
                name: "cron",
                file: "cron.js",
                fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
            }
        },
    },

    // Older runs with fixed ids and times (a future nextRun stays in the future during a test run), so the log has more than one page (size 10) and two scripts.
    // The traits name equals tags.name, as in the backend.
    jobRun('7001', 'every', everyTags, '2025-05-01T09:00:00Z', {
        schedule: '1m',
        maxRuns: 10,
        runs: 3,
        nextRun: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
        logs: [
            { level: 'info', message: 'sync started' },
            { level: 'debug', message: 'fetched 12 items from https://mokapi.io/api' },
        ],
    }),
    jobRun('7002', 'every', everyTags, '2025-05-01T08:59:00Z', {
        schedule: '1m',
        maxRuns: 10,
        runs: 2,
        nextRun: '2025-05-01T09:00:00Z',
        logs: null,
        error: { message: 'connection refused' },
    }),
    jobRun('7003', 'cron', cronTags, '2025-05-01T08:58:00Z', {
        schedule: '0 * * * *',
        maxRuns: 0,
        runs: 4,
        nextRun: '0001-01-01T00:00:00Z',
        logs: [{ level: 'info', message: 'last run' }],
    }),
    jobRun('7004', 'cleanup', cleanupTags, '2025-05-01T08:57:00Z', {
        schedule: '24h',
        maxRuns: -1,
        runs: 1,
        nextRun: '2025-05-02T08:57:00Z',
        logs: null,
    }),
    ...[6, 5, 4, 3, 2, 1, 0].map((minute, i) => {
        const name = i % 2 === 0 ? 'cron' : 'every'
        return jobRun(`${7005 + i}`, name, name === 'cron' ? cronTags : everyTags, `2025-05-01T08:${String(50 + minute).padStart(2, '0')}:00Z`, {
            schedule: name === 'cron' ? '3s' : '1m',
            maxRuns: -1,
            runs: 10 + i,
            nextRun: '2025-05-01T09:00:00Z',
            logs: [{ level: 'debug', message: `run ${10 + i}` }],
        })
    }),
]

const inFuture = (ms: number) => new Date(Date.now() + ms).toISOString()

// The result of the run event with this id, as the backend reports it in `lastRun`.
function lastRun(eventId: string, status: 'success' | 'warning' | 'error') {
    const event = jobs.find(x => x.id === eventId)!
    return { eventId, time: event.time, duration: event.data.duration, status }
}

// The registered jobs of `GET /api/jobs`. The names, files and Last Run ids match the run events above.
// In the order of the backend: sorted by name, then file; within a script by job index.
// The state `finished` and `cancelled` have no `nextRun`, a job that did not run yet has no `lastRun`.
//   cron.js   0 cron     cron `0 * * * *`, 4 runs, no limit, next run in 1 hour, success
//             1 cron     every 3s, 2 / 2 runs, finished, last run with an error
//             2 cron     every 1h, 1 run, cancelled, last run with a warning
//             3 cleanup  every 30m, not run yet, next run in 30 minutes (same name as the job of every.js)
//   every.js  0 every    every 1m, 3 / 10 runs, next run in 2 hours, success
//             1 cleanup  every 24h, 1 run, custom tag team, success
//             2 sync     every 10m, 1 run, the event of the run could not be stored (lastRun without eventId)
export const jobList = [
    {
        id: `${cronTags.fileKey}-3`,
        name: 'cleanup',
        schedule: '30m',
        maxRuns: -1,
        runs: 0,
        state: 'scheduled',
        nextRun: inFuture(30 * 60 * 1000),
        tags: { ...cronTags, name: 'cleanup' },
    },
    {
        id: `${everyTags.fileKey}-1`,
        name: 'cleanup',
        schedule: '24h',
        maxRuns: -1,
        runs: 1,
        state: 'scheduled',
        nextRun: inFuture(24 * 60 * 60 * 1000),
        tags: cleanupTags,
        lastRun: lastRun('7004', 'success'),
    },
    {
        id: `${cronTags.fileKey}-0`,
        name: 'cron',
        schedule: '0 * * * *',
        maxRuns: -1,
        runs: 4,
        state: 'scheduled',
        nextRun: inFuture(60 * 60 * 1000),
        tags: cronTags,
        lastRun: lastRun('7003', 'success'),
    },
    {
        id: `${cronTags.fileKey}-1`,
        name: 'cron',
        schedule: '3s',
        maxRuns: 2,
        runs: 2,
        state: 'finished',
        tags: cronTags,
        lastRun: lastRun('8134', 'error'),
    },
    {
        id: `${cronTags.fileKey}-2`,
        name: 'cron',
        schedule: '1h',
        maxRuns: -1,
        runs: 1,
        state: 'cancelled',
        tags: cronTags,
        lastRun: lastRun('81344', 'warning'),
    },
    {
        id: `${everyTags.fileKey}-0`,
        name: 'every',
        schedule: '1m',
        maxRuns: 10,
        runs: 3,
        state: 'scheduled',
        nextRun: inFuture(2 * 60 * 60 * 1000),
        tags: everyTags,
        lastRun: lastRun('7001', 'success'),
    },
    {
        id: `${everyTags.fileKey}-2`,
        name: 'sync',
        schedule: '10m',
        maxRuns: -1,
        runs: 1,
        state: 'scheduled',
        nextRun: inFuture(10 * 60 * 1000),
        tags: { ...everyTags, name: 'sync' },
        lastRun: { eventId: '', time: '2025-05-01T09:10:00Z', duration: 40, status: 'success' },
    },
]

export const config = {
    '6f5cab53-9f51-4bbe-bb32-fe41ee71d542': {
        id: '6f5cab53-9f51-4bbe-bb32-fe41ee71d542',
        url: 'file://cron.js',
        provider: 'file',
        time: '2025-05-01T08:49:25.482366+01:00',
        data: "import { every } from 'mokapi'; export default function() { every('1s', function() {}); }",
        filename: 'cron.js'
    },
    '0c3a5d52-7d6e-4b0f-8a3e-4c7a6d9b1e21': {
        id: '0c3a5d52-7d6e-4b0f-8a3e-4c7a6d9b1e21',
        url: 'file://every.js',
        provider: 'file',
        time: '2025-05-01T08:49:25.482366+01:00',
        data: "import { every } from 'mokapi'; export default function() { every('1m', function() {}); }",
        filename: 'every.js'
    }
}
