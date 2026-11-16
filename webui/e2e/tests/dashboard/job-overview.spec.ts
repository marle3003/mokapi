import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'
import { JobLogModel } from '../models/job'
import { JobOverviewModel } from '../models/job-overview'
import { useDashboardTabs } from '../components/dashboard'
import { pauseClock } from '../helpers/clock'
import { countdownSeconds } from '../helpers/countdown'
import { formatDateTime } from '../helpers/format'

// The jobs of the mock (e2e/mocks/jobs.ts, `jobList`) in the order of the API (the backend sorts them by name,
// then file); the overview shows them as sent:
//   0 cleanup  cron.js   30m         0 runs, scheduled, no last run (same name as row 1)
//   1 cleanup  every.js  24h         1, scheduled, last run 7004 success
//   2 cron     cron.js   0 * * * *   4, scheduled, next run in 1 hour, last run 7003 success
//   3 cron     cron.js   3s          2 / 2, finished, last run 8134 error
//   4 cron     cron.js   1h          1, cancelled, last run 81344 warning
//   5 every    every.js  1m          3 / 10, scheduled, next run in 2 hours, last run 7001 success
const cronKey = '6f5cab53-9f51-4bbe-bb32-fe41ee71d542'
const everyKey = '0c3a5d52-7d6e-4b0f-8a3e-4c7a6d9b1e21'
const names = ['cleanup', 'cleanup', 'cron', 'cron', 'cron', 'every']
const files = ['cron.js', 'every.js', 'cron.js', 'cron.js', 'cron.js', 'every.js']
const schedules = ['30m', '24h', '0 * * * *', '3s', '1h', '1m']
const states = ['Scheduled', 'Scheduled', 'Scheduled', 'Finished', 'Cancelled', 'Scheduled']
const runs = ['0', '1', '4', '2 / 2', '1', '3 / 10']
const cleanupOfCron = 0
const cleanupOfEvery = 1
const scheduledCron = 2
const finishedCron = 3
const cancelledCron = 4
const everyJob = 5
// `2026-11-16 12:00:00 (in 1h 2min 3s)`
const dateWithCountdown = /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \(in /

const jobsPath = '/dashboard/jobs'

function useJobsPage(page: Page) {
    return { ...layoutLocators(page), overview: new JobOverviewModel(page), log: new JobLogModel(page) }
}

// The response of the mock API, changed by the test.
async function serveJobs(page: Page, change: () => (jobs: JobInfo[]) => JobInfo[]) {
    await page.route('**/api/jobs', async route => {
        const response = await route.fetch()
        return route.fulfill({ response, json: change()(await response.json()) })
    })
}

test.use({ layout: 'new' })

test.describe('Job overview: tabs', () => {
    test('opens the overview on the Jobs tab', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.jobsTab).toHaveAttribute('aria-selected', 'true')
        await expect(overview.runsTab).toHaveAttribute('aria-selected', 'false')
        await expect(overview.table).toBeVisible()
        await expect(log.log).toHaveCount(0)
    })

    test('#tab-runs opens the log', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(`${jobsPath}#tab-runs`)

        await expect(overview.runsTab).toHaveAttribute('aria-selected', 'true')
        await expect(log.rows).toHaveCount(14)
        await expect(overview.table).toHaveCount(0)
    })

    test('the tabs switch the view and set the hash', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(jobsPath)

        await overview.runsTab.click()
        await expect(page).toHaveURL(/#tab-runs$/)
        await expect(log.rows).toHaveCount(14)

        await overview.jobsTab.click()
        await expect(page).toHaveURL(/#tab-jobs$/)
        await expect(overview.rows).toHaveCount(6)
        await expect(log.log).toHaveCount(0)
    })

    test('a selected run (?job=) opens Runs with its panel', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(`${jobsPath}?job=8134`)

        await expect(overview.runsTab).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.element).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('the hash wins over a selected run', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(`${jobsPath}?job=8134#tab-jobs`)

        await expect(overview.jobsTab).toHaveAttribute('aria-selected', 'true')
        await expect(overview.table).toBeVisible()
    })

    test('switching to Jobs drops the selected run', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(`${jobsPath}?job=8134`)

        await overview.jobsTab.click()

        await expect(overview.table).toBeVisible()
        await expect(page).not.toHaveURL(/job=/)
    })
})

test.describe('Job overview: table', () => {
    test('shows the columns', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.columnHeaders).toHaveText(['Name', 'Schedule', 'State', 'Runs', 'Next Run', 'Last Run', 'File'])
    })

    test('lists one row per job in the order of the API', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.rows).toHaveCount(6)
        await expect(overview.column('Name')).toHaveText(names)
        await expect(overview.column('File')).toHaveText(files)
        await expect(overview.column('Schedule')).toHaveText(schedules)
    })

    test('shows the state of a job', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.column('State')).toHaveText(states)
    })

    test('shows the runs, with the limit if there is one', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.column('Runs')).toHaveText(runs)
    })

    test('a job that did not run yet is listed without a last run', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        const row = overview.rows.nth(cleanupOfCron)
        await expect(overview.cell(row, 'Runs')).toHaveText('0')
        await expect(overview.cell(row, 'Last Run')).toHaveText('-')
    })

    test('the file links to the config of the script', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.cell(overview.rows.nth(scheduledCron), 'File').getByRole('link', { name: 'cron.js' }))
            .toHaveAttribute('href', `/dashboard/config/${cronKey}`)
        await expect(overview.cell(overview.rows.nth(everyJob), 'File').getByRole('link', { name: 'every.js' }))
            .toHaveAttribute('href', `/dashboard/config/${everyKey}`)
    })

    test('the file link opens the config', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await overview.cell(overview.rows.nth(everyJob), 'File').getByRole('link', { name: 'every.js' }).click()

        await expect(page).toHaveURL(new RegExp(`/dashboard/config/${everyKey}`))
    })
})

test.describe('Job overview: Next Run', () => {
    test('shows the date with a countdown while it is in the future', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.cell(overview.rows.nth(scheduledCron), 'Next Run')).toHaveText(dateWithCountdown)
        await expect(overview.cell(overview.rows.nth(everyJob), 'Next Run')).toHaveText(dateWithCountdown)
    })

    test('the countdown ticks every second', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.clock.install()
        await page.goto(jobsPath)
        const nextRun = overview.cell(overview.rows.nth(scheduledCron), 'Next Run')
        await expect(nextRun).toContainText('(in ')
        await pauseClock(page)
        const before = countdownSeconds(await nextRun.textContent())

        await page.clock.runFor(3000)

        await expect.poll(async () => countdownSeconds(await nextRun.textContent())).toBe(before - 3)
    })

    test('shows a dash for a finished and a cancelled job', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.cell(overview.rows.nth(finishedCron), 'Next Run')).toHaveText('-')
        await expect(overview.cell(overview.rows.nth(cancelledCron), 'Next Run')).toHaveText('-')
    })

    test('shows the plain date when the next run is in the past', async ({ page }) => {
        const { overview } = useJobsPage(page)
        const past = '2025-05-01T09:00:00Z'
        await serveJobs(page, () => jobs => jobs.map(job => job.name === 'every' ? { ...job, nextRun: past } : job))
        await page.goto(jobsPath)

        await expect(overview.cell(overview.rows.nth(everyJob), 'Next Run')).toHaveText(formatDateTime(past))
    })
})

test.describe('Job overview: Last Run', () => {
    test('shows the status and the time and links to the run', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        const lastRun = overview.cell(overview.rows.nth(everyJob), 'Last Run')
        await expect(lastRun).toContainText('Success')
        await expect(lastRun.getByRole('link')).toHaveText(formatDateTime('2025-05-01T09:00:00Z'))
        await expect(lastRun.getByRole('link')).toHaveAttribute('href', '/dashboard/jobs/7001')
    })

    test('shows the status Success, Warning and Error', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await expect(overview.cell(overview.rows.nth(cleanupOfEvery), 'Last Run')).toContainText('Success')
        await expect(overview.cell(overview.rows.nth(cancelledCron), 'Last Run')).toContainText('Warning')
        await expect(overview.cell(overview.rows.nth(finishedCron), 'Last Run')).toContainText('Error')
    })

    test('the time opens the full view of the run', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.goto(jobsPath)

        await overview.cell(overview.rows.nth(scheduledCron), 'Last Run').getByRole('link').click()

        await expect(page).toHaveURL(/\/dashboard\/jobs\/7003$/)
    })
})

test.describe('Job overview: row click', () => {
    test('opens Runs with the name of the job as chip', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(jobsPath)

        await overview.cell(overview.rows.nth(scheduledCron), 'Name').click()

        await expect(overview.runsTab).toHaveAttribute('aria-selected', 'true')
        await expect(page).toHaveURL(/#tab-runs$/)
        await expect(log.chips).toHaveText(['Name: cron'])
        await expect(log.rows).toHaveCount(8)
    })

    test('a click anywhere in the row works', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(jobsPath)

        await overview.cell(overview.rows.nth(everyJob), 'State').click()

        await expect(log.chips).toHaveText(['Name: every'])
        await expect(log.rows).toHaveCount(5)
    })

    test('a click on a link does not open Runs', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto(jobsPath)

        await overview.cell(overview.rows.nth(cleanupOfEvery), 'Last Run').getByRole('link').click()

        await expect(page).toHaveURL(/\/dashboard\/jobs\/7004$/)
        await expect(log.log).toHaveCount(0)
    })
})

test.describe('Job overview: no jobs', () => {
    test('shows No jobs defined with links to the docs', async ({ page }) => {
        const { overview } = useJobsPage(page)
        await page.route('**/api/jobs', route => route.fulfill({ json: [] }))
        await page.goto(jobsPath)

        await expect(overview.emptyState).toBeVisible()
        await expect(overview.rows).toHaveCount(0)
        await expect(page.getByRole('link', { name: 'mokapi.every' })).toHaveAttribute('href', '/docs/javascript-api/mokapi/every')
        await expect(page.getByRole('link', { name: 'mokapi.cron' })).toHaveAttribute('href', '/docs/javascript-api/mokapi/cron')
    })

    test('the Activity Bar shows Jobs without a job metric', async ({ page }) => {
        const { activityBar, overview } = useJobsPage(page)
        await page.route('**/api/metrics/app', route => route.fulfill({ json: [] }))
        await page.route('**/api/jobs', route => route.fulfill({ json: [] }))
        await page.goto('/dashboard')

        await activityBar.getByRole('link', { name: 'Jobs' }).click()

        await expect(page).toHaveURL(/\/dashboard\/jobs/)
        await expect(overview.emptyState).toBeVisible()
    })
})

test.describe('Job overview: refresh', () => {
    test('reloads the jobs with ?refresh while the countdown keeps ticking', async ({ page }) => {
        const { overview } = useJobsPage(page)
        let added = false
        await serveJobs(page, () => jobs => added
            ? [...jobs, { ...jobs[0], id: `${everyKey}-2`, name: 'zeta', tags: { ...jobs[0].tags, name: 'zeta' } }]
            : jobs)
        await page.clock.install()
        await page.goto(`${jobsPath}?refresh=1`)
        await expect(overview.rows).toHaveCount(6)
        await pauseClock(page)
        const nextRun = overview.cell(overview.rows.nth(scheduledCron), 'Next Run')
        const before = countdownSeconds(await nextRun.textContent())

        added = true
        await page.clock.runFor(1000)

        await expect(overview.rows).toHaveCount(7)
        await expect(overview.column('Name').last()).toHaveText('zeta')
        await expect.poll(async () => countdownSeconds(await nextRun.textContent())).toBe(before - 1)
    })
})

test.describe('Job overview: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('the Jobs tab shows the overview and Runs the log', async ({ page }) => {
        const { overview, log } = useJobsPage(page)
        await page.goto('/dashboard')
        await useDashboardTabs(page).jobs.click()

        await expect(overview.rows).toHaveCount(6)

        await overview.runsTab.click()

        await expect(log.rows).toHaveCount(14)
    })

    test('the tab bar still hides Jobs when no job ran', async ({ page }) => {
        const tabs = useDashboardTabs(page)
        await page.route('**/api/metrics/app', route => route.fulfill({ json: [] }))
        await page.goto('/dashboard')

        await expect(tabs.configs).toBeVisible()
        await expect(tabs.jobs).toHaveCount(0)
    })
})
