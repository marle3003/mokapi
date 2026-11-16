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

export const jobTabs: JobTab[] = ['Logs', 'Details']
