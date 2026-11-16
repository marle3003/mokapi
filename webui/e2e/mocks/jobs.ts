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
