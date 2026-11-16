export type JobTab = 'Logs' | 'Details'
export type JobStatus = 'Success' | 'Warning' | 'Error'

// The status of a run: an error of the job or a log with the level error is an Error, a log with the
// level warn a Warning. The backend sends `logs: null` for a run without logs.
export function jobStatus(data: JobExecution): JobStatus {
    const levels = (data.logs ?? []).map(log => log.level)
    if (data.error || levels.includes('error')) {
        return 'Error'
    }
    return levels.includes('warn') ? 'Warning' : 'Success'
}

// The status of the last run in `GET /api/jobs` (computed by the backend with the same rule) as the label of a run.
export function jobStatusLabel(status: JobLastRun['status']): JobStatus {
    return { success: 'Success', warning: 'Warning', error: 'Error' }[status] as JobStatus
}

export type JobsTab = 'tab-jobs' | 'tab-runs'

// The hash selects the tab. Without a (known) hash a selected run (`?job=`) opens Runs, so a shared
// link lands on the panel.
export function activeJobsTab(hash: string, query: { job?: unknown }): JobsTab {
    if (hash === '#tab-runs' || hash === '#tab-jobs') {
        return hash.slice(1) as JobsTab
    }
    return query.job ? 'tab-runs' : 'tab-jobs'
}

// `3 / 10` for a job with a limit, else the runs only (the backend sends -1 or 0 for unlimited).
export function runsLabel(runs: number, maxRuns: number | undefined): string {
    return maxRuns !== undefined && maxRuns > 0 ? `${runs} / ${maxRuns}` : `${runs}`
}

// A job that is not registered anymore has the zero time of Go (year 1) as its next run.
export function hasNextRun(nextRun: string | undefined): boolean {
    return !!nextRun && new Date(nextRun).getUTCFullYear() > 1
}

// `1h 2min 3s`: units without a value are left out, except the seconds.
export function formatCountdown(ms: number): string {
    const totalSeconds = Math.floor(ms / 1000)
    const hours = Math.floor(totalSeconds / 3600)
    const minutes = Math.floor(totalSeconds / 60) % 60
    const seconds = totalSeconds % 60
    return [hours > 0 ? `${hours}h` : '', minutes > 0 ? `${minutes}min` : '', `${seconds}s`]
        .filter(Boolean)
        .join(' ')
}

// `-` without a next run, else the date; while it is in the future, a countdown shows next to it.
export function nextRunLabel(nextRun: string | undefined, now: number, formatDate: (date: string) => string): string {
    if (!nextRun || !hasNextRun(nextRun)) {
        return '-'
    }
    const date = formatDate(nextRun)
    const remaining = Date.parse(nextRun) - now
    return remaining > 0 ? `${date} (in ${formatCountdown(remaining)})` : date
}

// The state of a job in `GET /api/jobs` as shown in the overview.
export function jobStateLabel(state: JobInfo['state']): string {
    return { scheduled: 'Scheduled', finished: 'Finished', cancelled: 'Cancelled' }[state]
}

export const jobTabs: JobTab[] = ['Logs', 'Details']
