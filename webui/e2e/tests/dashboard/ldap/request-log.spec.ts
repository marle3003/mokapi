import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { LdapRequestLogModel } from '../../models/ldap'
import { useLdapRequests } from '../../components/ldap'
import { useRequestView } from '../../components/request-view'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

// The request events of the LDAP mock (e2e/mocks/ldap.ts) in the order of the log. LDAP Testserver has
// twelve, LDAP Second Server three:
//   search, modify, add, delete (failed, with matched DN), modifydn, compare (the original six),
//   then newId(1) bind, newId(2) unbind, newId(3) search with attributes, newId(4) search with Event Handlers,
//   newId(5) add (failed, with matched DN), newId(6) search without results,
//   and of the second server newId(7) bind, newId(8) search, newId(9) unbind.
const newId = (n: number) => `b1c1a001-0000-4000-8000-00000000000${n}`
const search = 'a1289b9b-aff7-4c53-92a0-808c7ce7d907'
const modify = 'fbe9ade9-2adc-451b-a86b-a900c06c0058'
const add = '762b4a9e-7f26-4314-8ffd-7ae1922ab330'
const remove = '23636b50-e19b-4acc-afc0-f5918a7d2e64'
const modifyDn = '2fa8df80-9b15-427b-9c6d-0503aec06ed3'
const compare = '34cbe69d-948c-43e3-aa19-df83c09116c4'
const bind = newId(1)
const unbind = newId(2)
const searchWithAttributes = newId(3)
const searchWithHandlers = newId(4)
const addFailed = newId(5)
const searchEmpty = newId(6)

const service = 'LDAP Testserver'
const secondService = 'LDAP Second Server'
const servicePath = encodeURIComponent(service)
const serviceLog = `/dashboard/ldap/log/${servicePath}`
const allServicesLog = '/dashboard/ldap/log'
const serviceView = `/dashboard/ldap/service/${servicePath}`
const requestPath = (id: string) => `/dashboard/ldap/ldap/requests/${id}`
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useRequestLog(page: Page) {
    return { ...layoutLocators(page), log: new LdapRequestLogModel(page), view: useRequestView(page) }
}

// More than one page: 10 rows per page, the log of all services has 15 requests.
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

test.use({ layout: 'new' })

test.describe('LDAP request log: routes', () => {
    test('the log of one service opens in a tab named after the service', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await expect(tabs).toHaveText([`${service} requests`])
        await expect(log.rows).toHaveCount(12)
        await expect(log.row(search)).toHaveCount(1)
        await expect(log.row(newId(7))).toHaveCount(0)
    })

    test('the log of all services shows the requests of every service', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(allServicesLog)

        await expect(tabs).toHaveText(['Requests'])
        await expect(log.rows).toHaveCount(15)
        await expect(log.row(search)).toHaveCount(1)
        await expect(log.row(newId(7))).toHaveCount(1)
    })

    test('a log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useRequestLog(page)
        await page.goto(serviceLog)

        const star = tabs.first().getByRole('button', { name: `Pin ${service} requests` })
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText([`${service} requests`])
        await expect(pinned.getByRole('link', { name: `${service} requests` })).toHaveAttribute('href', new RegExp(`/dashboard/ldap/log/${servicePath}$`))
    })

    test('the log of all services can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useRequestLog(page)
        await page.goto(allServicesLog)

        await tabs.first().getByRole('button', { name: 'Pin Requests' }).click()

        await expect(sidebar.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(['Requests'])
    })

    test('a selected request is not part of a pinned entry', async ({ page }) => {
        const { tabs, sidebar, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(bind)).click()
        await expect(page).toHaveURL(new RegExp(`request=${bind}`))

        await tabs.first().getByRole('button', { name: `Pin ${service} requests` }).click()

        const link = sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: `${service} requests` })
        await expect(link).toHaveAttribute('href', new RegExp(`/dashboard/ldap/log/${servicePath}$`))
    })

    test('selecting rows keeps one tab for the log', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(bind)).click()
        await expect(page).toHaveURL(new RegExp(`request=${bind}`))
        await log.time(log.row(unbind)).click()
        await expect(page).toHaveURL(new RegExp(`request=${unbind}`))

        await expect(tabs).toHaveText([`${service} requests`])
    })
})

test.describe('LDAP request log: Requests sub tab', () => {
    test('has the sub tabs Requests, Servers and Configs, Requests is selected', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceView)

        await expect(page.getByRole('tab')).toHaveText(['Requests', 'Servers', 'Configs'])
        await expect(page.getByRole('tab', { name: 'Requests', exact: true })).toHaveAttribute('class', /active/)
        await expect(log.rows).toHaveCount(12)
        await expect(page.getByRole('table', { name: 'Servers' })).toHaveCount(0)
    })

    test('Servers shows the servers and not the log', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceView)

        await page.getByRole('tab', { name: 'Servers', exact: true }).click()

        await expect(page).toHaveURL(/#tab-servers$/)
        await expect(page.getByRole('table', { name: 'Servers' })).toBeVisible()
        await expect(log.log).toHaveCount(0)
    })

    test('shows the log of the service and a link to open it as a tab', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceView)
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText(service)

        await expect(log.rows).toHaveCount(12)
        await expect(page.getByRole('link', { name: 'Open as tab' })).toHaveAttribute('href', new RegExp(`/dashboard/ldap/log/${servicePath}$`))
    })

    test('Open as tab replaces the tab with a plain click', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(`${serviceView}#tab-requests`)
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click()

        await expect(tabs).toHaveText([`${service} requests`])
        await expect(page).toHaveURL(new RegExp(`/dashboard/ldap/log/${servicePath}$`))
    })

    test('a log of a service can be opened as an additional tab', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(`${serviceView}#tab-requests`)
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click(newTab)

        await expect(tabs).toHaveText([service, `${service} requests`])
        await expect(activeTab).toHaveText(`${service} requests`)
    })

    test('Configs does not show the log', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceView}#tab-configs`)
        await expect(page.getByRole('tab', { name: 'Configs', exact: true })).toHaveAttribute('class', /active/)

        await expect(log.log).toHaveCount(0)
    })

    test('selecting a request and then switching the sub tab drops the selection', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceView}#tab-requests`)
        await log.time(log.row(bind)).click()
        await expect(page).toHaveURL(new RegExp(`request=${bind}`))

        await page.getByRole('tab', { name: 'Servers' }).click()

        await expect(page).not.toHaveURL(/request=/)
    })

    test('switching the sub tab keeps the refresh interval', async ({ page }) => {
        await page.goto(`${serviceView}?refresh=5#tab-requests`)

        await page.getByRole('tab', { name: 'Configs' }).click()

        await expect(page).toHaveURL(/refresh=5/)
        await expect(page).toHaveURL(/#tab-configs$/)
    })
})

test.describe('LDAP request log: protocol page', () => {
    test('keeps the log with panel, chips and Pause', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto('/dashboard/ldap?refresh=1')
        await expect(log.rows).toHaveCount(15)

        await log.time(log.rows.first()).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/request=/)
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.addFilter).toBeVisible()
        await expect(log.pause).toBeVisible()
    })

    test('the log stays when a filter has no result', async ({ page }) => {
        const { log } = useRequestLog(page)
        await seedStorage(page, 'ldap-requests-filter', { service: secondService, operation: 'modify' })
        await page.goto('/dashboard/ldap')

        await expect(log.log).toBeVisible()
        await expect(log.chips).toHaveText([`Service: ${secondService}`, 'Operation: Modify'])
        await expect(log.rows).toHaveCount(0)
    })
})

test.describe('LDAP request log: detail panel', () => {
    test('a click on a row opens the panel and highlights the row', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.panel.element).toHaveCount(0)

        await log.time(log.row(search)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/ldap/requests/${search}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`/dashboard/ldap/log/${servicePath}\\?request=${search}$`))
    })

    test('the row has one link, the operation, to the full request view', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        await expect(log.row(search).getByRole('link')).toHaveCount(1)
        await expect(log.row(search).getByRole('link')).toHaveText('Search')
        await expect(log.row(search).getByRole('link')).toHaveAttribute('href', requestPath(search))
    })

    test('a click on the operation of a row selects it instead of leaving the log', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.row(modify).getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`\\?request=${modify}$`))
    })

    // [name, id, operation, head text, status, time]
    const heads: [string, string, string, string | undefined, string, string][] = [
        ['Search', search, 'Search', '(objectClass=user)', 'Success', '2025-02-27T11:52:25.482366+01:00'],
        ['Add', add, 'Add', 'cn=alice,cn=users,dc=mokapi,dc=io', 'Success', '2025-02-27T11:50:25.482366+01:00'],
        ['Modify', modify, 'Modify', 'cn=turangal,cn=users,dc=mokapi,dc=io', 'Success', '2025-02-27T11:50:50.482366+01:00'],
        ['Delete', remove, 'Delete', 'cn=turangal,cn=users,dc=mokapi,dc=io', 'NoSuchObject', '2025-02-27T11:50:10.482366+01:00'],
        ['ModifyDN', modifyDn, 'ModifyDN', 'cn=turangal,cn=users,dc=mokapi,dc=io', 'Success', '2025-02-27T11:49:25.482366+01:00'],
        ['Compare', compare, 'Compare', 'cn=turangal,cn=users,dc=mokapi,dc=io', 'CompareTrue', '2025-02-27T11:49:10.482366+01:00'],
        ['Bind', bind, 'Bind', 'cn=admin,dc=mokapi,dc=io', 'Success', '2025-02-27T11:48:50.482366+01:00'],
        ['Unbind', unbind, 'Unbind', undefined, '-', '2025-02-27T11:48:40.482366+01:00'],
    ]
    for (const [name, id, operation, text, status, time] of heads) {
        test(`the head of a ${name} request shows the operation, ${text ? 'the DN or filter' : 'no text'}, the status and the time`, async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(serviceLog)

            await log.time(log.row(id)).click()

            const panel = log.panel.element
            const head = view.field(panel, 'Operation')
            await expect(head.locator('.badge')).toHaveText(operation)
            if (text) {
                await expect(head).toContainText(text)
            } else {
                await expect(head).toHaveText(operation)
            }
            await expect(view.field(panel, 'Status')).toHaveText(status)
            await expect(view.field(panel, 'Time')).toHaveText(formatDateTime(time))
        })
    }

    // [name, id, sub tabs]
    const tabsOfOperation: [string, string, string[]][] = [
        ['Bind', bind, ['Request', 'Details']],
        ['Search', search, ['Request', 'Response', 'Details']],
        ['Add', add, ['Request', 'Details']],
        ['Modify', modify, ['Request', 'Details']],
        ['ModifyDN', modifyDn, ['Request', 'Details']],
        ['Compare', compare, ['Request', 'Details']],
        ['Delete', remove, ['Details']],
        ['Unbind', unbind, ['Details']],
    ]
    for (const [name, id, tabs] of tabsOfOperation) {
        test(`a ${name} request has the sub tabs ${tabs.join(', ')}, the first is selected`, async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(serviceLog)

            await log.time(log.row(id)).click()

            await expect(log.panel.tabs).toHaveText(tabs)
            await expect(log.panel.tab(tabs[0])).toHaveAttribute('aria-selected', 'true')
        })
    }

    test('Request shows the fields of a Bind: name, auth and password', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(bind)).click()

        await expect(view.field(log.panel.content, 'Name')).toHaveText('cn=admin,dc=mokapi,dc=io')
        await expect(view.field(log.panel.content, 'Auth')).toHaveText('simple')
        await expect(view.field(log.panel.content, 'Password')).toHaveText('secret')
    })

    test('Request shows the fields of a Search', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(search)).click()

        const content = log.panel.content
        await expect(view.field(content, 'Base DN')).toHaveText('dc=mokapi,dc=io')
        await expect(view.field(content, 'Scope')).toHaveText('SingleLevel')
        await expect(view.field(content, 'Size Limit')).toHaveText('10')
        await expect(view.field(content, 'Time Limit')).toHaveText('30 [s]')
        await expect(view.field(content, 'Attributes')).toHaveText('memberOf, thumbnailphoto')
    })

    test('Request shows the changes of a Modify and the attributes of an Add', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(modify)).click()
        await view.expectRows(log.panel.content.getByRole('table'), [['Add', 'foo', 'bar']])

        await log.time(log.row(add)).click()
        await view.expectRows(log.panel.content.getByRole('table'), [['cn', 'alice'], ['objectClass', 'topperson'], ['sn', 'Smith']])
    })

    test('Request shows the fields of a ModifyDN and a Compare', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(modifyDn)).click()
        await expect(view.field(log.panel.content, 'New RDN')).toHaveText('cn=foo')
        await expect(view.field(log.panel.content, 'New Superior DN')).toHaveText('cn=bar,dc=mokapi,dc=io')

        await log.time(log.row(compare)).click()
        await expect(view.field(log.panel.content, 'Attribute')).toHaveText('foo')
        await expect(view.field(log.panel.content, 'Value')).toHaveText('bar')
    })

    test('Response lists the results of a Search', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.tab('Response').click()

        await expect(log.panel.tab('Response')).toHaveAttribute('aria-selected', 'true')
        await view.expectRows(log.panel.content.getByRole('table'), [
            ['cn=farnsworthh,cn=users,dc=mokapi,dc=io', '0'],
            ['cn=fryp,cn=users,dc=mokapi,dc=io', '0'],
            ['cn=turangal,cn=users,dc=mokapi,dc=io', '0'],
            ['cn=wonga,cn=users,dc=mokapi,dc=io', '0'],
        ])
    })

    test('Response of a Search without results is an empty table', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(searchEmpty)).click()

        await log.panel.tab('Response').click()

        await expect(log.panel.content.getByRole('table')).toBeVisible()
        await expect(log.panel.content.getByRole('table').locator('tbody tr')).toHaveCount(0)
    })

    test('Details shows the duration', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(view.field(log.panel.content, 'Duration')).toBeVisible()
    })

    test('Details shows the Event Handlers of a request that has some', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(searchWithHandlers)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByText('Event Handlers')).toBeVisible()
    })

    test('Details has no Event Handlers for a request without', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByText('Event Handlers')).toHaveCount(0)
    })

    test('Details shows the message and the matched DN of a failed Delete', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(remove)).click()

        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(view.field(log.panel.content, 'Message')).toHaveText('delete operation failed: the specified entry does not exist: cn=turangal,cn=users,dc=mokapi,dc=io')
        await expect(view.field(log.panel.content, 'Matched DN')).toHaveText('cn=users,dc=mokapi,dc=io')
    })

    test('Details shows the message and the matched DN of a failed Add', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(addFailed)).click()

        await log.panel.tab('Details').click()

        await expect(view.field(log.panel.content, 'Message')).toContainText('entry already exists')
        await expect(view.field(log.panel.content, 'Matched DN')).toHaveText('cn=users,dc=mokapi,dc=io')
    })

    for (const [name, id, tab] of [['Bind', bind, 'Details'], ['Unbind', unbind, 'Details']]) {
        test(`Details of a ${name} without message and matched DN shows neither`, async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(serviceLog)
            await log.time(log.row(id)).click()

            await log.panel.tab(tab).click()

            await expect(log.panel.content.getByText('Message', { exact: true })).toHaveCount(0)
            await expect(log.panel.content.getByText('Matched DN', { exact: true })).toHaveCount(0)
        })
    }

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.tab('Response').click()

        await expect(page).toHaveURL(new RegExp(`\\?request=${search}$`))
    })

    test('a sub tab the next request does not have falls back to the first tab', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()
        await log.panel.tab('Response').click()

        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`request=${modify}`))
        await expect(log.panel.tabs).toHaveText(['Request', 'Details'])
        await expect(log.panel.tab('Request')).toHaveAttribute('aria-selected', 'true')
    })

    test('a request without Request tab opens on Details', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(add)).click()
        await expect(log.panel.tab('Request')).toHaveAttribute('aria-selected', 'true')

        await log.time(log.row(remove)).click()

        await expect(log.panel.tabs).toHaveText(['Details'])
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
    })

    test('the selected sub tab stays for the next request that has it', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(modify)).click()
        await log.panel.tab('Details').click()

        await log.time(log.row(add)).click()

        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
    })

    test('Ctrl/Cmd+click on a row opens the full request in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(search)).click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('LDAP Request')
        await expect(page).toHaveURL(new RegExp(`/requests/${search}$`))
    })

    test('a middle click on a row opens the full request in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(search)).click({ button: 'middle' })

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('LDAP Request')
        await expect(page).toHaveURL(new RegExp(`/requests/${search}$`))
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['LDAP Request'])
        await expect(page).toHaveURL(new RegExp(`/requests/${search}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('LDAP Request')
        await expect(page).toHaveURL(new RegExp(`/requests/${search}$`))
    })

    test('the full view opens from a row with the helper', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)

        await useLdapRequests(page).openFullView(log.time(log.row(bind)))

        await expect(page).toHaveURL(new RegExp(`/requests/${bind}$`))
        await expect(view.field(view.head, 'Operation').locator('.badge')).toHaveText('Bind')
    })

    test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`request=${modify}`))
        await expect(log.selectedRow.locator(`a[href$="/ldap/requests/${modify}"]`)).toHaveCount(1)

        await page.keyboard.press('ArrowUp')
        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`request=${search}`))
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useRequestLog(page)
        await usePageSize10(page)
        await page.goto(allServicesLog)
        await expect(log.rows).toHaveCount(10)
        await log.time(log.row(searchWithHandlers)).click()
        await expect(page).toHaveURL(new RegExp(`request=${searchWithHandlers}`))

        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`request=${searchWithHandlers}`))
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('arrow keys do nothing while the panel is closed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        await page.keyboard.press('ArrowDown')

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).not.toHaveURL(/request=/)
    })

    test('the selection survives a reload', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(bind)).click()
        await expect(page).toHaveURL(new RegExp(`request=${bind}`))

        await page.reload()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Operation').locator('.badge')).toHaveText('Bind')
        await expect(log.selectedRow.locator(`a[href$="/ldap/requests/${bind}"]`)).toHaveCount(1)
    })

    test('a link with ?request= opens the panel', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(`${serviceLog}?request=${unbind}`)

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Operation').locator('.badge')).toHaveText('Unbind')
    })

    test('selections do not add browser history entries', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto('/dashboard/configs')
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()
        await expect(page).toHaveURL(new RegExp(`request=${search}`))
        await log.time(log.row(modify)).click()
        await expect(page).toHaveURL(new RegExp(`request=${modify}`))
        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`request=${add}`))

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a request that is gone shows a message in the panel', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?request=00000000-0000-4000-8000-000000000000`)

        await expect(log.panel.element.getByText('LDAP Request not found')).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    test('is shown below the table in a narrow window', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
    })

    test('is shown beside the table in a wide window', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(search)).click()
        await expect(log.panel.element).toBeVisible()

        const card = await log.log.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.x).toBeGreaterThanOrEqual(card!.x + card!.width - 1)
        expect(panel!.y).toBeLessThan(card!.y + card!.height)
    })
})

test.describe('LDAP request log: Search result dialog', () => {
    async function openResult(page: Page) {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?request=${searchWithAttributes}`)
        await log.panel.tab('Response').click()
        await log.panel.content.getByRole('row', { name: /cn=alice,cn=users/ }).click()
        return { log, dialog: page.getByRole('dialog') }
    }

    test('a click on a result opens the attributes dialog above the backdrop', async ({ page }) => {
        const { view } = useRequestLog(page)
        const { dialog } = await openResult(page)

        await expect(dialog).toBeVisible()
        await expect(dialog).toContainText('DN: cn=alice,cn=users,dc=mokapi,dc=io')
        await view.expectRows(dialog.getByRole('table'), [['cn', 'alice'], ['objectClass', 'topperson'], ['sn', 'Smith']])
        // the click only reaches the button when nothing, like the backdrop, covers it
        await dialog.getByRole('button', { name: 'Close' }).click()
        await expect(dialog).toHaveCount(0)
    })

    test('ArrowDown inside the dialog does not move the selection of the log', async ({ page }) => {
        const { dialog } = await openResult(page)
        await expect(dialog).toBeVisible()

        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`request=${searchWithAttributes}`))
    })

    test('Escape closes the dialog and keeps the panel', async ({ page }) => {
        const { log, dialog } = await openResult(page)
        await expect(dialog).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(dialog).toHaveCount(0)
        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`request=${searchWithAttributes}`))
    })
})

test.describe('LDAP request log: full request view', () => {
    const handler = {
        duration: 1500000,
        tags: { name: 'dashboard', file: 'examples/mokapi/ldap_handler.js', fileKey: 'c7d6b1a2-5f0e-4c4a-9d0a-1b3c8e7f2a10', event: 'ldap' },
        logs: [{ level: 'info', message: 'a custom log message from the ldap handler' }],
        parameters: [JSON.stringify({ operation: 'Bind' }), JSON.stringify({ status: 'Success' })],
    }

    // The mock has Event Handlers for one request only. The others get one on their way to the dashboard.
    async function serveWithHandler(page: Page, id: string) {
        await page.route(`**/api/events/${id}`, async route => {
            const response = await route.fetch()
            const body = await response.json()
            return route.fulfill({ response, json: { ...body, data: { ...body.data, actions: [handler] } } })
        })
    }

    test('the head shows the operation, status, time and duration', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(search))

        await expect(view.field(view.head, 'Operation').locator('.badge')).toHaveText('Search')
        await expect(view.field(view.head, 'Operation')).toContainText('(objectClass=user)')
        await expect(view.field(view.head, 'Status')).toHaveText('Success')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime('2025-02-27T11:52:25.482366+01:00'))
        await expect(view.field(view.head, 'Duration')).toBeVisible()
    })

    test('a Bind shows the name in the head and name, auth and password', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(bind))

        await expect(view.field(view.head, 'Operation')).toContainText('cn=admin,dc=mokapi,dc=io')
        await expect(view.field(view.request, 'Name')).toHaveText('cn=admin,dc=mokapi,dc=io')
        await expect(view.field(view.request, 'Auth')).toHaveText('simple')
        await expect(view.field(view.request, 'Password')).toHaveText('secret')
    })

    test('a Search shows Request, Event Handlers and Response in this order', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(searchWithHandlers))
        const handlers = page.getByRole('region', { name: 'Event Handlers' })
        await expect(handlers).toBeVisible()

        await expect(view.field(view.request, 'Base DN')).toHaveText('dc=mokapi,dc=io')
        const [request, event, response] = await Promise.all([view.request, handlers, view.response].map(x => x.boundingBox()))
        expect(request!.y).toBeLessThan(event!.y)
        expect(event!.y).toBeLessThan(response!.y)
    })

    test('a Search lists its results and opens the attributes dialog', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(searchWithAttributes))

        await view.expectRows(view.response.getByRole('table'), [
            ['cn=alice,cn=users,dc=mokapi,dc=io', '3'],
            ['cn=alice2,cn=users,dc=mokapi,dc=io', '3'],
        ])
        await view.response.getByRole('row', { name: /cn=alice2/ }).click()
        await expect(page.getByRole('dialog')).toContainText('DN: cn=alice2,cn=users,dc=mokapi,dc=io')
    })

    test('a Modify, an Add, a ModifyDN and a Compare show their fields', async ({ page }) => {
        const { view } = useRequestLog(page)

        await page.goto(requestPath(modify))
        await view.expectRows(view.request.getByRole('table'), [['Add', 'foo', 'bar']])

        await page.goto(requestPath(add))
        await view.expectRows(view.request.getByRole('table'), [['cn', 'alice'], ['objectClass', 'topperson'], ['sn', 'Smith']])

        await page.goto(requestPath(modifyDn))
        await expect(view.field(view.request, 'New RDN')).toHaveText('cn=foo')

        await page.goto(requestPath(compare))
        await expect(view.field(view.request, 'Attribute')).toHaveText('foo')
    })

    test('a Delete has no Request, its head shows the DN', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(remove))

        await expect(view.field(view.head, 'Operation')).toContainText('cn=turangal,cn=users,dc=mokapi,dc=io')
        await expect(view.request).toHaveCount(0)
    })

    test('the head of a failed Delete shows the message and the matched DN', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(remove))

        await expect(view.field(view.head, 'Message')).toHaveText('delete operation failed: the specified entry does not exist: cn=turangal,cn=users,dc=mokapi,dc=io')
        await expect(view.field(view.head, 'Matched DN')).toHaveText('cn=users,dc=mokapi,dc=io')
    })

    test('the head of a request without message and matched DN shows neither', async ({ page }) => {
        const { view } = useRequestLog(page)
        await page.goto(requestPath(bind))
        await expect(view.field(view.head, 'Status')).toHaveText('Success')

        await expect(view.field(view.head, 'Message')).toHaveCount(0)
        await expect(view.field(view.head, 'Matched DN')).toHaveCount(0)
    })

    const operations: [string, string][] = [['Bind', bind], ['Modify', modify], ['Add', add], ['Delete', remove], ['ModifyDN', modifyDn], ['Compare', compare]]
    for (const [name, id] of operations) {
        test(`the Event Handlers of a ${name} are headed "Event Handlers"`, async ({ page }) => {
            await serveWithHandler(page, id)
            await page.goto(requestPath(id))

            await expect(page.getByRole('region', { name: 'Event Handlers' })).toBeVisible()
            await expect(page.getByRole('heading', { name: 'Actions' })).toHaveCount(0)
        })
    }

    test('an unknown request shows a message', async ({ page }) => {
        await page.goto(requestPath('00000000-0000-4000-8000-000000000000'))

        await expect(page.getByText('LDAP Request not found')).toBeVisible()
    })
})

test.describe('LDAP request log: filter chips', () => {
    test('shows no chip without a filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        await expect(log.rows).toHaveCount(12)
        await expect(log.chips).toHaveCount(0)
        await expect(log.addFilter).toBeVisible()
    })

    test('the operation is chosen in the popover and shown as the protocol names it', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')

        await expect(popover).toHaveCount(0)
        await expect(log.chips).toHaveText(['Operation: Search'])
        await expect(log.rows).toHaveCount(4)
        await expect(log.row(searchWithHandlers)).toHaveCount(1)
    })

    test('the popover offers all eight operations', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Operation')

        await expect(popover.getByRole('combobox', { name: 'Operation' }).getByRole('option')).toHaveText(['', 'Bind', 'Unbind', 'Search', 'Add', 'Modify', 'Delete', 'ModifyDN', 'Compare'])
    })

    test('the chip of ModifyDN reads like the operation', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'ModifyDN' })
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Operation: ModifyDN'])
        await expect(log.rows).toHaveCount(1)
        await expect(log.row(modifyDn)).toHaveCount(1)
    })

    test('the service is chosen in the popover of the log of all services', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(allServicesLog)

        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(secondService)
        await page.keyboard.press('Escape')

        await expect(log.chip(`Service: ${secondService}`).element).toBeVisible()
        await expect(log.rows).toHaveCount(3)
        await expect(log.row(newId(7))).toHaveCount(1)
    })

    test('the operation filter applies to all services', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(allServicesLog)

        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Bind' })
        await page.keyboard.press('Escape')

        await expect(log.rows).toHaveCount(2)
        await expect(log.row(bind)).toHaveCount(1)
        await expect(log.row(newId(7))).toHaveCount(1)
    })

    const offered: [string, string, string[]][] = [
        ['the log of all services', allServicesLog, ['Service', 'Operation']],
        ['the log of a service', serviceLog, ['Operation']],
        ['the protocol page', '/dashboard/ldap', ['Service', 'Operation']],
    ]
    for (const [name, path, fields] of offered) {
        test(`${name} does not offer the fields the view fixes`, async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(path)
            await expect(log.rows.first()).toBeVisible()

            await log.addFilter.click()

            await expect(log.addFilterMenu.getByRole('menuitem')).toHaveText(fields)
        })
    }

    test('a click on a chip edits the filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(4)

        await log.chip('Operation: Search').edit.click()
        await page.getByRole('dialog', { name: 'Operation', exact: true }).getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Bind' })
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Operation: Bind'])
        await expect(log.rows).toHaveCount(1)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(4)

        await log.chip('Operation: Search').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('filters survive a reload', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(4)

        await page.reload()

        await expect(log.chips).toHaveText(['Operation: Search'])
        await expect(log.rows).toHaveCount(4)
    })

    test('every place of the log has its own filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText(['Operation: Search'])

        await page.goto(allServicesLog)
        await expect(log.rows).toHaveCount(15)
        await expect(log.chips).toHaveCount(0)

        await page.goto(`/dashboard/ldap/log/${encodeURIComponent(secondService)}`)
        await expect(log.rows).toHaveCount(3)
        await expect(log.chips).toHaveCount(0)
    })
})

// A request of LDAP Testserver, served on top of the log once `arrive()` was called.
const arrivingRequest = {
    id: newId(99),
    traits: { namespace: 'ldap', name: service, operation: 'bind' },
    time: '2025-02-28T09:30:00+01:00',
    data: {
        duration: 2,
        request: { operation: 'Bind', version: '3', name: 'cn=new,dc=mokapi,dc=io', password: 'secret', auth: 'simple' },
        response: { status: 'Success' },
    }
}

async function serveArrivingRequest(page: Page) {
    let arrived = false
    let loads = 0
    await page.route('**/api/events?*', async route => {
        loads++
        const response = await route.fetch()
        if (!arrived) {
            return route.fulfill({ response })
        }
        const body = await response.json()
        return route.fulfill({ response, json: { ...body, items: [arrivingRequest, ...body.items], total: body.total + 1 } })
    })
    return {
        arrive: () => { arrived = true },
        loads: () => loads,
    }
}

test.describe('LDAP request log: Pause', () => {
    // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
    // so only runFor() lets a reload happen.
    async function openRefreshingLog(page: Page) {
        const { log } = useRequestLog(page)
        const events = await serveArrivingRequest(page)
        await page.clock.install()
        await page.goto(`${serviceLog}?refresh=1`)
        await expect(log.rows).toHaveCount(12)
        await pauseClock(page)
        return { log, events }
    }

    test('a new request appears on top and the selection stays on the same request', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(bind)).click()
        await expect(log.selectedRow).toHaveCount(1)

        events.arrive()
        await page.clock.runFor(1000)

        await expect(log.rows).toHaveCount(13)
        await expect(log.rows.first().locator(`a[href$="/ldap/requests/${arrivingRequest.id}"]`)).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/ldap/requests/${bind}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`request=${bind}`))
    })

    test('the Pause button is on the right, as in the other logs', async ({ page }) => {
        const { log } = await openRefreshingLog(page)

        const card = await log.log.boundingBox()
        const pause = await log.pause.boundingBox()
        expect(pause!.x).toBeGreaterThan(card!.x + card!.width / 2)
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
        await expect(log.rows).toHaveCount(12)
    })

    test('Resume loads at once and keeps the selection', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(bind)).click()
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        events.arrive()
        await page.clock.runFor(3000)
        await expect(log.rows).toHaveCount(12)

        await log.resume.click()

        await expect(log.rows).toHaveCount(13)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
        await expect(log.selectedRow.locator(`a[href$="/ldap/requests/${bind}"]`)).toHaveCount(1)
    })

    test('no Pause without ?refresh', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)
        await expect(log.pause).toHaveCount(0)
        await expect(log.paused).toHaveCount(0)
    })

    test('the pause is not kept after a reload', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?refresh=1`)
        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await page.reload()

        await expect(log.rows).toHaveCount(12)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})

// The Classic layout is the default of the fixture.
test.describe('LDAP request log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a request can be selected', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        await log.time(log.row(search)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Operation').locator('.badge')).toHaveText('Search')
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('a filter can be added and removed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText(['Operation: Search'])
        await expect(log.rows).toHaveCount(4)

        await log.chip('Operation: Search').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?refresh=1`)
        await expect(log.rows).toHaveCount(12)

        await log.pause.click()
        await expect(log.paused).toBeVisible()
        await log.resume.click()

        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
