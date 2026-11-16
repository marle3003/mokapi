import { test, expect } from '../models/fixture-website'
import { JobLogModel } from '../models/job'
import { useRequestView } from '../components/request-view'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

// The recording has one run of kafka.ts (max runs 2, runs 1, no logs, no error).
test.describe('Job log in the demo dashboard', () => {
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    test('shows the recorded run', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs#tab-runs`)

        const log = new JobLogModel(page)
        await expect(log.rows).toHaveCount(1)
        await expect(log.rows.first().getByRole('cell')).toContainText(['Success', 'kafka.ts', '5m'])
    })

    // the ids of the events change with every recording, the row is found by its name
    test('a run opens in the panel with the hint on console.log() and its details', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs#tab-runs`)
        const log = new JobLogModel(page)
        const view = useRequestView(page)

        await log.time(log.rows.filter({ hasText: 'kafka.ts' })).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/job=/)
        await expect(log.panel.openAsTab).toBeVisible()
        await expect(log.panel.tabs).toHaveText(['Logs', 'Details'])
        await expect(log.panel.content.getByText('console.log()')).toBeVisible()

        await log.panel.tab('Details').click()
        await expect(view.field(log.panel.element, 'Runs')).toHaveText('1 / 2')
        await expect(view.field(log.panel.element, 'Schedule')).toHaveText('5m')
        await expect(log.panel.content.getByRole('table', { name: 'Tags' })).toContainText('random message generator')
    })

    // the recording is a few minutes old when the test runs, so the clock moves far away to make the next run past
    test('the next run is a plain date when it is in the past', async ({ page }) => {
        await page.clock.install({ time: new Date('2100-01-01T00:00:00Z') })
        await page.goto(`${dashboardPath}/jobs#tab-runs`)
        const log = new JobLogModel(page)
        const view = useRequestView(page)

        await log.time(log.rows.filter({ hasText: 'kafka.ts' })).click()
        await log.panel.tab('Details').click()

        await expect(view.field(log.panel.element, 'Next Run')).toHaveText(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
    })

    test('an unknown run id shows that the job is not found', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs?job=00000000-0000-4000-8000-000000000000`)

        const log = new JobLogModel(page)
        await expect(log.panel.element.getByText('Job not found')).toBeVisible()
    })

    test('the Name chip filters the log', async ({ page }) => {
        await page.goto(`${dashboardPath}/jobs#tab-runs`)
        const log = new JobLogModel(page)
        await expect(log.rows).toHaveCount(1)

        const popover = await log.openAddFilter('Name')
        await popover.locator('input').fill('kafka.ts')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Name: kafka.ts'])
        await expect(log.rows).toHaveCount(1)
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await page.goto(`${dashboardPath}/jobs#tab-runs`)

        const log = new JobLogModel(page)
        await expect(log.rows).toHaveCount(1)
        await expect(log.pause).toHaveCount(0)
    })
})
