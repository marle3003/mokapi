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

    }
]