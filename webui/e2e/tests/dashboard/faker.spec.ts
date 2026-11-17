import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'
import { FakerModel } from '../models/faker'

// The tree of the mock (e2e/mocks/faker.ts):
//   /name /email /city /tag
//   /person            name firstname (weight 1, depends on gender and sex) lastname
//   /contact           phone (weight 0.5) email (weight 0.5) type
//   /address           street zip
//   /frequency         custom
//   /billing           custom, groups only, child code (custom)
//   /shipping          custom, groups only, child code (custom)
// name and email exist in several branches, code under billing and shipping.
const fakerPath = '/dashboard/faker'
const firstLevel = ['/', '/name', '/email', '/city', '/tag', '/person', '/contact', '/address', '/frequency', '/billing', '/shipping']

test.use({ layout: 'new' })

test.describe('Faker tree', () => {
    test('opens from the Activity Bar with the first level expanded', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto('/dashboard')
        await layoutLocators(page).activityBar.getByRole('link', { name: 'Faker' }).click()

        await expect(page).toHaveURL(/\/dashboard\/faker/)
        await expect(faker.row('/email').element).toBeVisible()
        expect(await faker.nodePaths()).toEqual(firstLevel)
    })

    test('shows name, attributes and the custom marker, a group looks different', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)

        await expect(faker.row('/city').element).toContainText('city')
        await expect(faker.row('/city').element).toContainText('locality')
        await expect(faker.row('/frequency').customMarker).toBeVisible()
        await expect(faker.row('/city').customMarker).toHaveCount(0)
        await expect(faker.row('/billing').groupMarker).toBeVisible()
        await expect(faker.row('/city').groupMarker).toHaveCount(0)
    })

    test('opens and closes a node', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)
        await expect(faker.row('/person/name').element).toHaveCount(0)

        await faker.row('/person').toggle.click()
        await expect(faker.row('/person/name').element).toBeVisible()
        expect(await faker.nodePaths()).toContain('/person/lastname')

        await faker.row('/person').toggle.click()
        await expect(faker.row('/person/name').element).toHaveCount(0)
    })

    test('filter shows matches with their ancestors open', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)

        await faker.filter.fill('code')

        // zip has the attributes zipCode, postcode and postalCode
        await expect.poll(() => faker.nodePaths()).toEqual(['/', '/address', '/address/zip', '/billing', '/billing/code', '/shipping', '/shipping/code'])
    })

    test('filter matches attributes and ignores case', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)

        await faker.filter.fill('FULLNAME')

        await expect.poll(() => faker.nodePaths()).toEqual(['/', '/person', '/person/name'])
    })

    test('filter without a match says so', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)

        await faker.filter.fill('nothing like this')

        await expect(faker.noMatch).toBeVisible()
        expect(await faker.nodePaths()).toEqual([])
    })

    test('Custom only shows custom nodes with their ancestors', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)

        await faker.customOnly.check()

        await expect.poll(() => faker.nodePaths()).toEqual(['/', '/frequency', '/billing', '/billing/code', '/shipping', '/shipping/code'])
    })

    test('Custom only combines with the filter', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)

        await faker.customOnly.check()
        await faker.filter.fill('code')
        await expect.poll(() => faker.nodePaths()).toEqual(['/', '/billing', '/billing/code', '/shipping', '/shipping/code'])

        await faker.filter.fill('email')
        await expect(faker.noMatch).toBeVisible()
    })

    test('Reload loads the tree again and keeps selection and open nodes', async ({ page }) => {
        const faker = new FakerModel(page)
        let reloaded = false
        await page.route('**/api/faker/tree', async route => {
            const response = await route.fetch()
            const tree = await response.json()
            if (reloaded) {
                tree.children.push({ name: 'added', path: '/added', attributes: ['added'], custom: true, hasFakeFunc: true })
            }
            await route.fulfill({ response, json: tree })
        })
        await page.goto(`${fakerPath}?node=/contact/email`)
        await expect(faker.row('/contact/email').select).toBeVisible()
        await expect(faker.row('/added').element).toHaveCount(0)

        reloaded = true
        await faker.reload.click()

        await expect(faker.row('/added').element).toBeVisible()
        await expect(faker.row('/contact/email').select).toBeVisible()
        await expect(faker.detail.field('Path')).toHaveText('/contact/email')
    })
})

test.describe('Faker node detail', () => {
    test('selects a node in the address and shows its detail', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)
        await faker.row('/person').toggle.click()

        await faker.row('/person/firstname').select.click()

        expect(faker.selectedInAddress()).toBe('/person/firstname')
        await expect(faker.row('/person/firstname').select).toHaveAttribute('aria-current', 'true')
        await expect(faker.detail.field('Name')).toHaveText('firstname')
        await expect(faker.detail.field('Path')).toHaveText('/person/firstname')
        await expect(faker.detail.field('Attributes')).toHaveText('firstname, firstname2')
        await expect(faker.detail.field('Weight')).toHaveText('1')
        await expect(faker.detail.field('Depends on')).toHaveText('gender, sex')
        await expect(faker.detail.field('Custom')).toHaveText('No')
    })

    test('has no weight for a node without one', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(`${fakerPath}?node=/frequency`)

        await expect(faker.detail.field('Name')).toHaveText('frequency')
        await expect(faker.detail.field('Custom')).toHaveText('Yes')
        await expect(faker.detail.element.getByText('Weight', { exact: true })).toHaveCount(0)
    })

    test('restores the selection after a reload of the page and opens the tree to it', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(`${fakerPath}?node=/contact/email`)

        await expect(faker.detail.field('Path')).toHaveText('/contact/email')
        await expect(faker.row('/contact/email').select).toHaveAttribute('aria-current', 'true')

        await page.reload()

        await expect(faker.detail.field('Path')).toHaveText('/contact/email')
        await expect(faker.row('/contact/email').select).toHaveAttribute('aria-current', 'true')
    })

    test('selecting does not add history entries and keeps the filter', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)
        const historyLength = await page.evaluate(() => history.length)
        await faker.filter.fill('code')

        await faker.row('/billing/code').select.click()
        await faker.row('/shipping/code').select.click()

        expect(await page.evaluate(() => history.length)).toBe(historyLength)
        await expect(faker.filter).toHaveValue('code')
        expect(faker.selectedInAddress()).toBe('/shipping/code')
    })

    test('says so for an unknown node', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(`${fakerPath}?node=/does/not/exist`)

        await expect(faker.detail.notFound).toBeVisible()
        await expect(faker.row('/email').element).toBeVisible()
    })

    test('Generate produces a value of exactly that node when names are the same', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(`${fakerPath}?node=/billing/code`)

        await faker.detail.generate.click()
        await expect(faker.detail.value).toHaveText(/^"BILL-\d+"$/)
        const first = await faker.detail.value.textContent()

        await faker.detail.generate.click()
        await expect(faker.detail.value).not.toHaveText(first!)
        await expect(faker.detail.value).toHaveText(/^"BILL-\d+"$/)

        await faker.row('/shipping').toggle.click()
        await faker.row('/shipping/code').select.click()
        await expect(faker.detail.value).toHaveCount(0)
        await faker.detail.generate.click()
        await expect(faker.detail.value).toHaveText(/^"SHIP-\d+"$/)
    })

    test('has no Generate for a node that only groups', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(`${fakerPath}?node=/billing`)

        await expect(faker.detail.field('Path')).toHaveText('/billing')
        await expect(faker.detail.generate).toHaveCount(0)
    })
})

test.describe('Faker playground', () => {
    test('shows the node and tokens of a property path', async ({ page }) => {
        const { playground } = new FakerModel(page)
        await page.goto(fakerPath)

        await playground.path.fill('customerEmail')
        await playground.generate.click()

        await expect(playground.value).toHaveText('"jane.doe@example.com"')
        const row = playground.matches.getRow(1)
        await expect(row.getCellByName('Property path')).toHaveText('customerEmail')
        await expect(row.getCellByName('Node')).toHaveText('/email')
        await expect(row.getCellByName('Tokens')).toHaveText('email')
        await expect(row.getCellByName('Value')).toHaveText('"jane.doe@example.com"')
    })

    test('shows the fallback for a schema only', async ({ page }) => {
        const { playground } = new FakerModel(page)
        await page.goto(fakerPath)

        await playground.schema.fill('{"type": "string", "format": "email"}')
        await playground.generate.click()

        const row = playground.matches.getRow(1)
        await expect(row.getCellByName('Property path')).toHaveText('(value)')
        await expect(row.getCellByName('Node')).toHaveText('fakeBySchema')
    })

    test('lists one row per nested property of an object schema', async ({ page }) => {
        const { playground } = new FakerModel(page)
        await page.goto(fakerPath)

        await playground.path.fill('user')
        await playground.schema.fill(JSON.stringify({
            type: 'object',
            properties: {
                email: { type: 'string' },
                address: { type: 'object', properties: { city: { type: 'string' } } },
                status: { type: 'string', enum: ['active', 'blocked'] },
                tags: { type: 'array', items: { type: 'string' } },
                code: { type: 'string' },
            }
        }))
        await playground.generate.click()

        // the mock lists the rows sorted by property path
        await expect(playground.matches.locator('tbody tr')).toHaveCount(5)
        const cell = (row: number, column: string) => playground.matches.getRow(row).getCellByName(column)
        await expect(cell(1, 'Property path')).toHaveText('user.address.city')
        await expect(cell(1, 'Node')).toHaveText('/city')
        await expect(cell(2, 'Property path')).toHaveText('user.code')
        await expect(cell(2, 'Node')).toHaveText('fakeBySchema')
        await expect(cell(3, 'Property path')).toHaveText('user.email')
        await expect(cell(3, 'Node')).toHaveText('/email')
        await expect(cell(4, 'Property path')).toHaveText('user.status')
        await expect(cell(4, 'Node')).toHaveText(/^-\s*from enum$/)
        await expect(cell(5, 'Property path')).toHaveText('user.tags[]')
        await expect(cell(5, 'Node')).toHaveText('/tag')
    })

    test('selects the node of a result row and opens the tree to it', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)
        await expect(faker.row('/contact/email').element).toHaveCount(0)

        await faker.playground.path.fill('contact.email')
        await faker.playground.generate.click()
        await expect(faker.playground.matches.getRow(1).getCellByName('Tokens')).toHaveText('contact email')
        await faker.playground.nodeButton(1).click()

        expect(faker.selectedInAddress()).toBe('/contact/email')
        await expect(faker.detail.field('Path')).toHaveText('/contact/email')
        await expect(faker.row('/contact/email').select).toHaveAttribute('aria-current', 'true')
    })

    test('shows the selected node of a result row although the filter hides it', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto(fakerPath)
        await faker.filter.fill('code')

        await faker.playground.path.fill('contact.email')
        await faker.playground.generate.click()
        await faker.playground.nodeButton(1).click()

        await expect(faker.row('/contact/email').select).toBeVisible()
        await expect(faker.filter).toHaveValue('')
    })

    test('does not send a request for invalid JSON and says so next to the field', async ({ page }) => {
        const { playground } = new FakerModel(page)
        const requests: string[] = []
        page.on('request', request => {
            if (request.url().includes('/api/faker/fake')) {
                requests.push(request.url())
            }
        })
        await page.goto(fakerPath)

        await playground.schema.fill('{"type": ')
        await playground.generate.click()

        await expect(playground.schemaError).toBeVisible()
        expect(requests).toEqual([])
    })

    test('shows the error of the generator', async ({ page }) => {
        const { playground } = new FakerModel(page)
        await page.goto(fakerPath)

        await playground.schema.fill('{"type": "unknown"}')
        await playground.generate.click()

        await expect(playground.generatorError).toHaveText('unsupported schema: schema type=unknown')
        await expect(playground.matches).toHaveCount(0)
    })

    test('replaces the error by the next result', async ({ page }) => {
        const { playground } = new FakerModel(page)
        await page.goto(fakerPath)
        await playground.schema.fill('{"type": "unknown"}')
        await playground.generate.click()
        await expect(playground.generatorError).toBeVisible()

        await playground.schema.fill('')
        await playground.path.fill('email')
        await playground.generate.click()

        await expect(playground.generatorError).toHaveCount(0)
        await expect(playground.value).toHaveText('"jane.doe@example.com"')
    })
})

test.describe('Faker route', () => {
    test('/dashboard/tree redirects to /dashboard/faker and keeps query and hash', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.goto('/dashboard/tree?node=/contact/email&refresh=20#playground')

        await expect(page).toHaveURL(/\/dashboard\/faker\?/)
        const url = new URL(page.url())
        expect(url.searchParams.get('node')).toBe('/contact/email')
        expect(url.searchParams.get('refresh')).toBe('20')
        expect(url.hash).toBe('#playground')
        await expect(faker.detail.field('Path')).toHaveText('/contact/email')
    })

    test('selecting a node keeps one tab and the playground input', async ({ page }) => {
        const faker = new FakerModel(page)
        const { tabs } = layoutLocators(page)
        await page.goto(fakerPath)
        await faker.playground.path.fill('customerEmail')
        const tabCount = await tabs.count()

        await faker.row('/email').select.click()
        await faker.row('/city').select.click()

        await expect(faker.playground.path).toHaveValue('customerEmail')
        await expect(tabs).toHaveCount(tabCount)
    })
})

test.describe('Faker layout', () => {
    test('has the tree beside the playground in a wide window', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(fakerPath)

        const tree = await faker.tree.boundingBox()
        const playground = await faker.playground.element.boundingBox()

        expect(playground!.x).toBeGreaterThanOrEqual(tree!.x + tree!.width - 1)
    })

    test('stacks the columns in a narrow window with the tree first', async ({ page }) => {
        const faker = new FakerModel(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(fakerPath)

        const tree = await faker.tree.boundingBox()
        const playground = await faker.playground.element.boundingBox()
        const detail = await faker.detail.element.boundingBox()

        expect(playground!.y).toBeGreaterThanOrEqual(tree!.y + tree!.height - 1)
        expect(detail!.y).toBeGreaterThanOrEqual(playground!.y + playground!.height - 1)
    })
})
