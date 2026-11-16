import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'
import { JobLogModel } from '../models/job'
import { useJobs } from '../components/job'
import { useRequestView } from '../components/request-view'
import { pauseClock } from '../helpers/clock'
import { formatDateTime } from '../helpers/format'

// The job runs of the mock (e2e/mocks/jobs.ts) in the order of the log. There are 14 runs of three jobs:
//   8234  cron     no logs, a next run in 1 hour (countdown), runs 1 without limit
//   8134  cron     an error with an error log (with a link), runs 1 / 10
//   81344 cron     a warning (with a link)
//   7001  every    info and debug logs (with a link), runs 3 / 10, a next run in 2 hours
//   7002  every    an error and `logs: null`, a past next run
//   7003  cron     the zero time as next run, no limit (0)
//   7004  cleanup  custom tags (team), `logs: null`
//   7005 to 7011 more runs of cron (7005, 7007, ...) and every (7006, 7008, ...) with a debug log, past next run
// cron has 8 runs, every 5 and cleanup 1.
const cron = { tags: { name: 'cron', file: 'cron.js', fileKey: '6f5cab53-9f51-4bbe-bb32-fe41ee71d542' } }
// the tags are sorted by name, as the mock API returns them
const cronTagRows = [['file', 'cron.js'], ['fileKey', cron.tags.fileKey], ['name', 'cron']]
const plain = '8234'
const errorRun = '8134'
const warningRun = '81344'
const everyRun = { id: '7001', time: '2025-05-01T09:00:00Z' }
const everyError = { id: '7002', time: '2025-05-01T08:59:00Z' }
const zeroNextRun = '7003'
const customTags = '7004'
const pastNextRun = { id: '7005', nextRun: '2025-05-01T09:00:00Z' }

const jobsPath = '/dashboard/jobs'
const jobPath = (id: string) => `/dashboard/jobs/${id}`
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useJobLog(page: Page) {
    return { ...layoutLocators(page), log: new JobLogModel(page), view: useRequestView(page) }
}

// More than one page: 10 rows per page, the log has 14 runs.
async function usePageSize10(page: Page) {
    await page.addInitScript(() => localStorage.setItem('page_size', '10'))
}

// A value of localStorage only on the first load of the tab, so a reload shows what the dashboard saved.
async function seedStorage(page: Page, key: string, value: unknown) {
    await page.addInitScript(([k, v]) => {
        if (!sessionStorage.getItem(`seeded-${k}`)) {
            sessionStorage.setItem(`seeded-${k}`, 'true')
            localStorage.setItem(k, v)
        }
    }, [key, JSON.stringify(value)])
}

// Vue ignores a click that is older than the handler. Under the paused clock an element that appears
// after the pause (the tabs of the panel) has a handler that is "attached" at the time of the pause, so
// the clock must move before the first click on it.
async function moveClockPastNewHandlers(page: Page) {
    await page.clock.runFor(10)
}

// `2026-11-16 12:00:00 (in 1h 2min 3s)` in seconds
function countdownSeconds(text: string | null): number {
    const match = /\(in (?:(\d+)h )?(?:(\d+)min )?(\d+)s\)/.exec(text ?? '')
    expect(match, `a countdown in "${text}"`).not.toBeNull()
    return Number(match![1] ?? 0) * 3600 + Number(match![2] ?? 0) * 60 + Number(match![3])
}

test.use({ layout: 'new' })

test.describe('Job log: page', () => {
    test('shows the log with paging, chip, Pause and no collapse rows', async ({ page }) => {
        const { log } = useJobLog(page)
        await usePageSize10(page)
        await page.goto(jobsPath)

        await expect(log.rows).toHaveCount(10)
        await expect(log.log.getByText('Showing 10 of 14 requests')).toBeVisible()
        await expect(log.addFilter).toBeVisible()
        await expect(log.pause).toBeVisible()
        await expect(log.table.locator('.collapse-row')).toHaveCount(0)
        await expect(log.table.locator('[data-bs-toggle="collapse"]')).toHaveCount(0)
    })

    test('the second page has the remaining runs', async ({ page }) => {
        const { log } = useJobLog(page)
        await usePageSize10(page)
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(10)

        await log.log.getByLabel('Next').click()

        await expect(log.rows).toHaveCount(4)
    })

    test('the log stays when a filter has no result', async ({ page }) => {
        const { log } = useJobLog(page)
        await seedStorage(page, 'job-jobs-filter', { name: 'unknown' })
        await page.goto(jobsPath)

        await expect(log.log).toBeVisible()
        await expect(log.chips).toHaveText(['Name: unknown'])
        await expect(log.rows).toHaveCount(0)
    })

    test('the log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useJobLog(page)
        await page.goto(jobsPath)

        await tabs.first().getByRole('button', { name: 'Pin Jobs' }).click()

        await expect(sidebar.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(['Jobs'])
    })
})

test.describe('Job log: table', () => {
    test('has the columns Status, Name, Schedule, Time and Duration and no Next Run', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await expect(log.table.getByRole('columnheader')).toHaveText(['Status', 'Name', 'Schedule', 'Time', 'Duration'])
    })

    test('a run shows status, name, schedule, time and duration', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await expect(log.row(everyRun.id).getByRole('cell')).toHaveText(['Success', 'every', '1m', formatDateTime(everyRun.time), '1 [sec]'])
        await expect(log.row(everyError.id).getByRole('cell')).toHaveText(['Error', 'every', '1m', formatDateTime(everyError.time), '1 [sec]'])
        await expect(log.row(plain).getByRole('cell').nth(4)).toHaveText('2 [sec]')
    })

    test('the status is Success, Warning or Error', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await expect(log.row(plain).locator('.badge')).toHaveText('Success')
        await expect(log.row(warningRun).locator('.badge')).toHaveText('Warning')
        await expect(log.row(errorRun).locator('.badge')).toHaveText('Error')
        await expect(log.row(everyError.id).locator('.badge')).toHaveText('Error')
    })

    test('a row has exactly one link, the name, to the full view', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await expect(log.row(plain).getByRole('link')).toHaveCount(1)
        await expect(log.row(plain).getByRole('link')).toHaveText('cron')
        await expect(log.row(plain).getByRole('link')).toHaveAttribute('href', jobPath(plain))
    })
})

test.describe('Job log: detail panel', () => {
    test('a click on a row opens the panel and highlights the row', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await expect(log.panel.element).toHaveCount(0)

        await log.time(log.row(errorRun)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/jobs/${errorRun}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`${jobsPath}\\?job=${errorRun}$`))
    })

    test('a click on the name selects the row instead of leaving the log', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.row(errorRun).getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`\\?job=${errorRun}$`))
    })

    test('Ctrl/Cmd+click on a row opens the full view in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(errorRun)).click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('Job')
        await expect(page).toHaveURL(new RegExp(`/jobs/${errorRun}$`))
    })

    test('a middle click on a row opens the full view in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(errorRun)).click({ button: 'middle' })

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('Job')
        await expect(page).toHaveURL(new RegExp(`/jobs/${errorRun}$`))
    })

    test('the head shows status, name and time', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(everyError.id)).click()

        const panel = log.panel.element
        await expect(view.field(panel, 'Status')).toHaveText('Error')
        await expect(view.field(panel, 'Name')).toHaveText('every')
        await expect(view.field(panel, 'Time')).toHaveText(formatDateTime(everyError.time))
    })

    test('has the sub tabs Logs and Details, Logs is selected', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(plain)).click()

        await expect(log.panel.tabs).toHaveText(['Logs', 'Details'])
        await expect(log.panel.tab('Logs')).toHaveAttribute('aria-selected', 'true')
    })

    test('Logs shows the error on top, the levels as icons and links in messages', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(errorRun)).click()

        const content = log.panel.content
        const alert = content.getByRole('alert')
        await expect(alert).toContainText('There was an error')
        await expect(content.locator('.bi-x-circle-fill')).toHaveCount(1)
        await expect(content.getByRole('link', { name: 'https://mokapi.io' })).toHaveAttribute('href', 'https://mokapi.io')
        const alertBox = await alert.boundingBox()
        const logBox = await content.locator('.list-group').boundingBox()
        expect(alertBox!.y).toBeLessThan(logBox!.y)
    })

    test('Logs shows a warning with its icon and link', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(warningRun)).click()

        const content = log.panel.content
        await expect(content.locator('.bi-exclamation-triangle-fill')).toHaveCount(1)
        await expect(content.getByText('a warning message, see')).toBeVisible()
        await expect(content.getByRole('alert')).toHaveCount(0)
    })

    test('Logs shows info and debug messages with their icons', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(everyRun.id)).click()

        const content = log.panel.content
        await expect(content.locator('.list-group-item')).toHaveText(['sync started', 'fetched 12 items from https://mokapi.io/api'])
        await expect(content.locator('.bi-chat-dots')).toHaveCount(1)
        await expect(content.locator('.bi-bug-fill')).toHaveCount(1)
    })

    test('Logs shows the hint on console.log() for a run without logs', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(plain)).click()
        await expect(log.panel.content.getByText('This job did not produce any logs')).toBeVisible()
        await expect(log.panel.content.getByText('console.log()')).toBeVisible()

        await log.time(log.row(customTags)).click()
        await expect(log.panel.content.getByText('This job did not produce any logs')).toBeVisible()
    })

    test('Logs shows the error of a run with `logs: null`', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(everyError.id)).click()

        await expect(log.panel.content.getByRole('alert')).toContainText('connection refused')
        await expect(log.panel.content.getByText('This job did not produce any logs')).toBeVisible()
    })

    test('Details shows schedule, runs with the limit, duration and next run', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(everyRun.id)).click()
        await log.panel.tab('Details').click()

        const panel = log.panel.element
        await expect(view.field(panel, 'Schedule')).toHaveText('1m')
        await expect(view.field(panel, 'Runs')).toHaveText('3 / 10')
        await expect(view.field(panel, 'Duration')).toHaveText('1 [sec]')
        await expect(view.field(panel, 'Next Run')).toHaveText(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \(in 1h \d+min \d+s\)$/)
    })

    test('Details shows the runs without a limit as a number', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(plain)).click()
        await log.panel.tab('Details').click()
        await expect(view.field(log.panel.element, 'Runs')).toHaveText('1')

        await log.time(log.row(zeroNextRun)).click()
        await expect(view.field(log.panel.element, 'Runs')).toHaveText('4')
    })

    test('Details shows the plain date for a next run in the past', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(pastNextRun.id)).click()
        await log.panel.tab('Details').click()

        await expect(view.field(log.panel.element, 'Next Run')).toHaveText(formatDateTime(pastNextRun.nextRun))
    })

    test('Details shows - for a job that has no next run', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(zeroNextRun)).click()
        await log.panel.tab('Details').click()

        await expect(view.field(log.panel.element, 'Next Run')).toHaveText('-')
    })

    test('the countdown of the next run ticks every second', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.clock.install()
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(14)
        await pauseClock(page)
        await log.time(log.row(plain)).click()
        await expect(log.panel.tab('Details')).toBeVisible()
        await moveClockPastNewHandlers(page)
        await log.panel.tab('Details').click()
        const nextRun = view.field(log.panel.element, 'Next Run')
        await expect(nextRun).toContainText('(in ')
        const before = countdownSeconds(await nextRun.textContent())

        await page.clock.runFor(3000)

        await expect.poll(async () => countdownSeconds(await nextRun.textContent())).toBe(before - 3)
    })

    test('Details shows the tags and the script file links to the config', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(plain)).click()
        await log.panel.tab('Details').click()

        const tags = log.panel.element.getByRole('table', { name: 'Tags' })
        await view.expectRows(tags, cronTagRows)
        await expect(tags.getByRole('link', { name: 'cron.js' })).toHaveAttribute('href', `/dashboard/config/${cron.tags.fileKey}`)
    })

    test('Details shows the custom tags of a job', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(customTags)).click()
        await log.panel.tab('Details').click()

        await expect(view.field(log.panel.element, 'Name')).toHaveText('cleanup')
        const tags = log.panel.element.getByRole('table', { name: 'Tags' })
        await expect(tags.locator('tbody tr').filter({ hasText: 'team' }).getByRole('cell')).toHaveText(['team', 'platform'])
    })

    test('the selected sub tab stays for the next run', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()
        await log.panel.tab('Details').click()

        await page.keyboard.press('ArrowDown')

        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(view.field(log.panel.element, 'Runs')).toHaveText('1 / 10')
    })

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()

        await log.panel.tab('Details').click()

        await expect(page).toHaveURL(new RegExp(`\\?job=${plain}$`))
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(errorRun)).click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['Job'])
        await expect(page).toHaveURL(new RegExp(`/jobs/${errorRun}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(errorRun)).click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('Job')
        await expect(page).toHaveURL(new RegExp(`/jobs/${errorRun}$`))
    })

    test('the full view opens from a row with the helper', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)

        await useJobs(page).openFullView(log.time(log.row(everyRun.id)))

        await expect(page).toHaveURL(new RegExp(`/jobs/${everyRun.id}$`))
        await expect(view.field(page.getByRole('main'), 'Name')).toHaveText('every')
    })

    test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`job=${errorRun}`))
        await expect(log.selectedRow.locator(`a[href$="/jobs/${errorRun}"]`)).toHaveCount(1)

        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`job=${plain}`))
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useJobLog(page)
        await usePageSize10(page)
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(10)
        await log.time(log.rows.last()).click()
        await expect(page).toHaveURL(/job=/)
        const url = page.url()

        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(url)
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${jobsPath}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${jobsPath}$`))
    })

    test('the selection survives a reload', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)
        await log.time(log.row(everyRun.id)).click()
        await expect(page).toHaveURL(new RegExp(`job=${everyRun.id}`))

        await page.reload()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Name')).toHaveText('every')
        await expect(log.selectedRow.locator(`a[href$="/jobs/${everyRun.id}"]`)).toHaveCount(1)
    })

    test('selections do not add browser history entries', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto('/dashboard/configs')
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()
        await expect(page).toHaveURL(new RegExp(`job=${plain}`))
        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`job=${errorRun}`))

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('the selection is not part of the tab identity', async ({ page }) => {
        const { tabs, log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.time(log.row(plain)).click()
        await expect(page).toHaveURL(new RegExp(`job=${plain}`))
        await log.time(log.row(errorRun)).click()
        await expect(page).toHaveURL(new RegExp(`job=${errorRun}`))

        await expect(tabs).toHaveText(['Jobs'])
    })

    test('a run that does not exist shows a message in the panel', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(`${jobsPath}?job=unknown`)

        await expect(log.panel.element.getByText('Job not found')).toBeVisible()
        await expect(log.rows).toHaveCount(14)
    })

    test('is shown below the table in a narrow window', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
    })

    test('is shown beside the table in a wide window', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(jobsPath)
        await log.time(log.row(plain)).click()
        await expect(log.panel.element).toBeVisible()

        const card = await log.log.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.x).toBeGreaterThanOrEqual(card!.x + card!.width - 1)
        expect(panel!.y).toBeLessThan(card!.y + card!.height)
    })
})

test.describe('Job log: full view', () => {
    test('shows head, error, logs, details and tags', async ({ page }) => {
        const { view } = useJobLog(page)
        await page.goto(jobPath(errorRun))

        const main = page.getByRole('main')
        await expect(view.field(main, 'Status')).toHaveText('Error')
        await expect(view.field(main, 'Name')).toHaveText('cron')
        await expect(main.getByRole('alert')).toContainText('There was an error')
        await expect(main.getByText('an error message, see')).toBeVisible()
        await expect(view.field(main, 'Schedule')).toHaveText('3s')
        await expect(view.field(main, 'Runs')).toHaveText('1 / 10')
        await expect(view.field(main, 'Duration')).toHaveText('2 [sec]')
        await expect(view.field(main, 'Next Run')).toHaveText(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
        const tags = main.getByRole('table', { name: 'Tags' })
        await view.expectRows(tags, cronTagRows)
        await expect(tags.getByRole('link', { name: 'cron.js' })).toHaveAttribute('href', `/dashboard/config/${cron.tags.fileKey}`)
    })

    test('shows the hint on console.log() for a run without logs', async ({ page }) => {
        await page.goto(jobPath(plain))

        await expect(page.getByRole('main').getByText('This job did not produce any logs')).toBeVisible()
    })

    test('an unknown run shows a message', async ({ page }) => {
        await page.goto(jobPath('unknown'))

        await expect(page.getByText('Job not found')).toBeVisible()
    })

    test('the tab is named Job and can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useJobLog(page)
        await page.goto(jobPath(errorRun))
        await expect(tabs).toHaveText(['Job'])

        await tabs.first().getByRole('button', { name: 'Pin Job' }).click()

        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText(['Job'])
        await expect(pinned.getByRole('link', { name: 'Job' })).toHaveAttribute('href', new RegExp(`${jobPath(errorRun)}$`))
    })

    test('the Jobs entry of the activity bar stays highlighted', async ({ page }) => {
        const { activityBar } = useJobLog(page)
        await page.goto(jobPath(errorRun))

        await expect(activityBar.getByRole('link', { name: 'Jobs' })).toHaveClass(/router-link-active/)
    })
})

test.describe('Job log: filter chips', () => {
    test('shows no chip without a filter', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await expect(log.rows).toHaveCount(14)
        await expect(log.chips).toHaveCount(0)
        await expect(log.addFilter).toBeVisible()
    })

    test('the name is entered in the popover', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        const popover = await log.openAddFilter('Name')
        await popover.locator('input').fill('every')
        await page.keyboard.press('Escape')

        await expect(log.chip('Name: every').element).toBeVisible()
        await expect(log.rows).toHaveCount(5)
        await expect(log.table.getByRole('cell', { name: 'every', exact: true })).toHaveCount(5)
    })

    test('the popover suggests the names of the loaded rows', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(14)

        const popover = await log.openAddFilter('Name')

        const names = await popover.locator('datalist option').evaluateAll(options => options.map(o => (o as HTMLOptionElement).value))
        expect(names).toEqual(['cleanup', 'cron', 'every'])
    })

    test('the suggestions follow the loaded rows', async ({ page }) => {
        const { log } = useJobLog(page)
        await seedStorage(page, 'job-jobs-filter', { name: 'cron' })
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(8)

        await log.chip('Name: cron').edit.click()

        const names = await page.getByRole('dialog', { name: 'Name', exact: true }).locator('datalist option')
            .evaluateAll(options => options.map(o => (o as HTMLOptionElement).value))
        expect(names).toEqual(['cron'])
    })

    test('a click on a chip edits the filter', async ({ page }) => {
        const { log } = useJobLog(page)
        await seedStorage(page, 'job-jobs-filter', { name: 'every' })
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(5)

        await log.chip('Name: every').edit.click()
        await page.getByRole('dialog', { name: 'Name', exact: true }).locator('input').fill('cleanup')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Name: cleanup'])
        await expect(log.rows).toHaveCount(1)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useJobLog(page)
        await seedStorage(page, 'job-jobs-filter', { name: 'every' })
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(5)

        await log.chip('Name: every').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(14)
    })

    test('the filter is saved and survives a reload', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        const popover = await log.openAddFilter('Name')
        await popover.locator('input').fill('every')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(5)

        await page.reload()

        await expect(log.chips).toHaveText(['Name: every'])
        await expect(log.rows).toHaveCount(5)
        expect(await page.evaluate(() => localStorage.getItem('job-jobs-filter'))).toBe(JSON.stringify({ name: 'every' }))
    })

    test('the Name chip is the only filter on offer', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)

        await log.addFilter.click()

        await expect(log.addFilterMenu.getByRole('menuitem')).toHaveText(['Name'])
    })
})

const arrivingJob = {
    id: '9001',
    traits: { namespace: 'job', name: 'cron' },
    time: new Date().toISOString(),
    data: { schedule: '3s', maxRuns: -1, runs: 11, nextRun: '2025-05-01T09:00:00Z', duration: 1000, logs: [], ...cron },
}

async function serveArrivingJob(page: Page) {
    let arrived = false
    let loads = 0
    await page.route('**/api/events?*', async route => {
        loads++
        const response = await route.fetch()
        if (!arrived) {
            return route.fulfill({ response })
        }
        const body = await response.json()
        return route.fulfill({ response, json: { ...body, items: [arrivingJob, ...body.items], total: body.total + 1 } })
    })
    return {
        arrive: () => { arrived = true },
        loads: () => loads,
    }
}

test.describe('Job log: Pause', () => {
    // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
    // so only runFor() lets a reload happen.
    async function openRefreshingLog(page: Page) {
        const { log } = useJobLog(page)
        const events = await serveArrivingJob(page)
        await page.clock.install()
        await page.goto(`${jobsPath}?refresh=1`)
        await expect(log.rows).toHaveCount(14)
        await pauseClock(page)
        return { log, events }
    }

    test('a new run appears on top and the selection stays on the same run', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(errorRun)).click()
        await expect(log.selectedRow).toHaveCount(1)

        events.arrive()
        await page.clock.runFor(1000)

        await expect(log.rows).toHaveCount(15)
        await expect(log.rows.first().locator(`a[href$="/jobs/${arrivingJob.id}"]`)).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/jobs/${errorRun}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`job=${errorRun}`))
    })

    test('Pause stops reloading and shows Paused', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await expect(log.paused).toHaveCount(0)

        await log.pause.click()

        await expect(log.paused).toBeVisible()
        await expect(log.resume).toBeVisible()
        await expect(log.pause).toHaveCount(0)
        const before = events.loads()
        events.arrive()
        await page.clock.runFor(3000)
        expect(events.loads()).toBe(before)
        await expect(log.rows).toHaveCount(14)
    })

    test('Resume loads at once and keeps the selection', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(errorRun)).click()
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        events.arrive()
        await page.clock.runFor(3000)
        await expect(log.rows).toHaveCount(14)

        await log.resume.click()

        await expect(log.rows).toHaveCount(15)
        await expect(log.paused).toHaveCount(0)
        await expect(log.selectedRow.locator(`a[href$="/jobs/${errorRun}"]`)).toHaveCount(1)
    })

    test('the countdown keeps running while the log is paused', async ({ page }) => {
        const { log } = await openRefreshingLog(page)
        await log.time(log.row(plain)).click()
        await expect(log.panel.tab('Details')).toBeVisible()
        await moveClockPastNewHandlers(page)
        await log.panel.tab('Details').click()
        const nextRun = log.panel.element.locator('p.label', { hasText: /^Next Run$/ }).locator('xpath=following-sibling::*[1]')
        await expect(nextRun).toContainText('(in ')
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        const before = countdownSeconds(await nextRun.textContent())

        await page.clock.runFor(3000)

        await expect.poll(async () => countdownSeconds(await nextRun.textContent())).toBe(before - 3)
    })

    test('the pause is not kept after a reload', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await page.reload()

        await expect(log.rows).toHaveCount(14)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})

// The Classic layout is the default of the fixture.
test.describe('Job log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a run can be selected', async ({ page }) => {
        const { log, view } = useJobLog(page)
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(14)

        await log.time(log.row(errorRun)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Name')).toHaveText('cron')
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useJobLog(page)
        await page.goto(jobsPath)
        await expect(log.rows).toHaveCount(14)

        await log.pause.click()
        await expect(log.paused).toBeVisible()
        await log.resume.click()

        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })

    test('the Jobs tab is active on the full view', async ({ page, dashboard }) => {
        await page.goto(jobPath(errorRun))

        await expect(dashboard.tabs.getByRole('link', { name: 'Jobs', exact: true })).toHaveClass(/router-link-active/)
        await expect(dashboard.tabs.getByRole('link', { name: 'Configs', exact: true })).not.toHaveClass(/router-link-active/)
    })
})
