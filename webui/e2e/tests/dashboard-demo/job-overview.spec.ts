import { test, expect } from '../models/fixture-website'
import { JobLogModel } from '../models/job'
import { JobOverviewModel } from '../models/job-overview'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

// The recording has one job, kafka.ts (every 5m, max runs 2), with one run.
test.describe('Job overview in the demo dashboard', () => {
    let dashboardPath: string

    test.beforeEach(({ baseURL }) => {
        dashboardPath = baseURL !== 'http://localhost:8080' ? '/dashboard-demo' : '/dashboard'
    })

    test('lists the job of kafka.ts', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs`)

        const overview = new JobOverviewModel(page)
        await expect(overview.rows).toHaveCount(1)
        const row = overview.rows.first()
        await expect(overview.cell(row, 'Schedule')).toHaveText('5m')
        await expect(overview.cell(row, 'State')).toHaveText('Scheduled')
        await expect(overview.cell(row, 'Runs')).toHaveText('1 / 2')
        await expect(overview.cell(row, 'File')).toContainText('kafka.ts')
    })

    // the ids of the events change with every recording, so only the target is checked
    test('the last run links to the run', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs`)

        const overview = new JobOverviewModel(page)
        const row = overview.rows.first()
        await expect(overview.cell(row, 'Last Run')).toContainText('Error')
        await expect(overview.cell(row, 'Last Run').getByRole('link')).toHaveAttribute('href', /\/jobs\/[\w-]+$/)
    })

    // the recording is a few minutes old when the test runs, so the clock moves far away to make the next run past
    test('the next run is a plain date when it is in the past', async ({ page }) => {
        await page.clock.install({ time: new Date('2100-01-01T00:00:00Z') })
        await page.goto(`${dashboardPath}/jobs`)

        const overview = new JobOverviewModel(page)
        await expect(overview.cell(overview.rows.first(), 'Next Run')).toHaveText(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
    })

    test('a click on the row shows the runs of the job', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs`)

        const overview = new JobOverviewModel(page)
        await overview.cell(overview.rows.first(), 'Name').click()

        await expect(overview.runsTab).toHaveAttribute('aria-selected', 'true')
        await expect(page).toHaveURL(/#tab-runs$/)
        const log = new JobLogModel(page)
        await expect(log.chips).toHaveText(['Name: kafka.ts'])
        await expect(log.rows).toHaveCount(1)
    })
})
