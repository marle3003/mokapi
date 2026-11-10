import { expect, test, type Page } from '@playwright/test'

// The website is pre-rendered: every page is served as static HTML with its content and head
// (see ssg/prerender.js), the app then hydrates it. The checks need the files of a build, so they
// run on the preview build only (CI set). `vite dev` renders in the browser.
test.skip(!process.env.CI, 'needs the pre-rendered pages of a build, run with CI=true')

// `/` is the shell of the app, the pre-rendered home page is served as /home
const pages = [
    {
        path: '/home',
        h1: 'The Open-Source Mock API Tool Across Protocols',
        title: 'Mokapi: Open-Source Mock API Tool & API Simulation',
        description: 'Simulate HTTP, Kafka, MQTT, WebSocket, LDAP and mail APIs locally. Open-source mock API tool driven by OpenAPI and AsyncAPI specs.',
        canonical: 'https://mokapi.io',
    },
    {
        path: '/http',
        h1: 'OpenAPI Mock Server & HTTP API Simulation',
        title: 'OpenAPI Mock Server & HTTP REST API Mocking | Mokapi',
        description: 'Spin up an open-source HTTP mock server instantly from your OpenAPI specifications. Validate REST requests, simulate edge cases, and automate your testing.',
        canonical: 'https://mokapi.io/http',
    },
    {
        path: '/docs/http/overview',
        h1: 'Mocking HTTP APIs',
        title: 'How to mock HTTP APIs with Mokapi',
        description: 'Mock any HTTP API with OpenAPI specification',
        canonical: 'https://mokapi.io/docs/http/overview',
    },
]

function decodeEntities(text: string): string {
    return text
        .replace(/&lt;/g, '<')
        .replace(/&gt;/g, '>')
        .replace(/&quot;/g, '"')
        .replace(/&#39;/g, "'")
        .replace(/&amp;/g, '&')
}

function attributeOf(tag: string | undefined, attribute: string): string | undefined {
    const value = tag?.match(new RegExp(`\\s${attribute}="([^"]*)"`))?.[1]
    return value === undefined ? undefined : decodeEntities(value)
}

function metaContent(html: string, key: 'name' | 'property', name: string): string | undefined {
    const tag = html.match(new RegExp(`<meta\\s[^>]*${key}="${name}"[^>]*>`))?.[0]
    return attributeOf(tag, 'content')
}

function canonicalOf(html: string): string | undefined {
    return attributeOf(html.match(/<link\s[^>]*rel="canonical"[^>]*>/)?.[0], 'href')
}

function titleOf(html: string): string | undefined {
    const title = html.match(/<title>([^<]*)<\/title>/)?.[1]
    return title === undefined ? undefined : decodeEntities(title)
}

function h1Of(html: string): string | undefined {
    const h1 = html.match(/<h1[^>]*>([\s\S]*?)<\/h1>/)?.[1]
    return h1 === undefined ? undefined : decodeEntities(h1.replace(/<[^>]*>/g, '').trim())
}

test.describe('pre-rendered HTML', () => {

    for (const page of pages) {
        test(`${page.path} contains its content and head without running JavaScript`, async ({ request }) => {
            const response = await request.get(page.path)
            expect(response.status()).toBe(200)
            const html = await response.text()

            expect(h1Of(html)).toBe(page.h1)
            expect(titleOf(html)).toContain(page.title)
            expect(metaContent(html, 'name', 'description')).toBe(page.description)
            expect(canonicalOf(html)).toBe(page.canonical)
            expect(metaContent(html, 'property', 'og:title')).toContain(page.title)
            expect(metaContent(html, 'property', 'og:description')).toBe(page.description)
            expect(metaContent(html, 'property', 'og:url')).toBe(page.canonical)
            expect(metaContent(html, 'property', 'og:image')).toBe('https://mokapi.io/og-logo.png')
            expect(metaContent(html, 'property', 'og:type')).toBe('website')
        })
    }

    // not in sitemap.xml, the prerender finds them through the links of other pages
    for (const page of [
        { path: '/resources', h1: 'Mokapi Resources', title: 'Explore Mokapi Resources' },
        { path: '/docs', h1: 'Mocking APIs with Mokapi', title: 'Mokapi Documentation' },
    ]) {
        test(`${page.path} is pre-rendered although it is not in the sitemap`, async ({ request }) => {
            const html = await (await request.get(page.path)).text()

            const app = html.match(/<div id="app">([\s\S]*)<\/div>/)?.[1] ?? ''
            expect(h1Of(app)).toBe(page.h1)
            expect(titleOf(html)).toContain(page.title)
        })
    }

    test('/home contains the structured data of the software', async ({ request }) => {
        const html = await (await request.get('/home')).text()

        const json = html.match(/<script type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/)?.[1]
        expect(json, 'JSON-LD script').toBeDefined()
        expect(JSON.parse(json!)).toMatchObject({ '@type': 'SoftwareApplication', name: 'Mokapi', url: 'https://mokapi.io' })
    })

    test('/http contains highlighted code in the landing page code block', async ({ request }) => {
        const html = await (await request.get('/http')).text()

        expect(html).toMatch(/<pre[^>]*>\s*<code[^>]*>(?:(?!<\/code>)[\s\S])*<span class="hljs-/)
    })
})

test.describe('without JavaScript', () => {
    test.use({ javaScriptEnabled: false })

    test('a docs page shows its heading', async ({ page }) => {
        await page.goto('/docs/http/overview')

        await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs', level: 1 })).toBeVisible()
    })
})

test.describe('hydration', () => {

    // Vue 3 production builds don't warn about mismatches. When the client doesn't adopt the
    // pre-rendered DOM it renders it again and the nodes are replaced, which this test catches.
    // Vue always sets the HTML of a v-html element again (the Markdown of a docs page), so for docs
    // pages the element around it is checked.
    for (const { path, selector } of [
        { path: '/http', selector: 'h1' },
        { path: '/docs/http/overview', selector: 'main .content' },
    ]) {
        test(`${path} keeps the pre-rendered DOM instead of rendering it again`, async ({ page }) => {
            const releaseScripts = await holdScripts(page)

            try {
                await page.goto(path, { waitUntil: 'commit' })
                const element = page.locator(selector)
                await expect(element).toBeVisible()
                expect(await isAppMounted(page)).toBe(false)
                const prerendered = await element.elementHandle()

                releaseScripts()
                await expectAppMounted(page)

                expect(await prerendered!.evaluate((node, selector) => node === document.querySelector(selector), selector)).toBe(true)
            } finally {
                releaseScripts()
            }
        })
    }

    test('the page works after hydration: theme switch', async ({ page }) => {
        await page.goto('/docs/http/overview')
        await expectAppMounted(page)
        await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')

        await visibleThemeIcon(page, 'Switch to light mode').click()

        await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')
        await expect(visibleThemeIcon(page, 'Switch to dark mode')).toBeVisible()
    })

    // The project stores the dark theme in localStorage. The page is pre-rendered once for all
    // visitors, so the icon must follow the theme the inline script of the page sets.
    test('the header shows the icon of the stored theme before and after hydration', async ({ page }) => {
        const releaseScripts = await holdScripts(page)

        try {
            await page.goto('/docs/http/overview', { waitUntil: 'commit' })
            await expect(visibleThemeIcon(page, 'Switch to light mode')).toBeVisible()
            await expect(visibleThemeIcon(page, 'Switch to dark mode')).toHaveCount(0)
            expect(await isAppMounted(page)).toBe(false)

            releaseScripts()
            await expectAppMounted(page)

            await expect(visibleThemeIcon(page, 'Switch to light mode')).toBeVisible()
            await expect(visibleThemeIcon(page, 'Switch to dark mode')).toHaveCount(0)
        } finally {
            releaseScripts()
        }
    })
})

// Vue marks the container of a mounted app, also when it hydrates (data-v-app is only set when the app
// renders on its own)
async function isAppMounted(page: Page): Promise<boolean> {
    return page.evaluate(() => '__vue_app__' in document.getElementById('app')!)
}

async function expectAppMounted(page: Page) {
    await page.waitForFunction(() => '__vue_app__' in document.getElementById('app')!)
}

// Keeps the JavaScript of the app back, so the page can be looked at as the server rendered it
async function holdScripts(page: Page): Promise<() => void> {
    let release!: () => void
    const released = new Promise<void>((resolve) => { release = resolve })
    await page.route(/\/assets\/[^/]+\.js$/, async (route) => {
        await released
        await route.continue().catch(() => { /* test finished */ })
    })
    return release
}

// The header has the tools twice (desktop and collapsed navbar), only one is visible
function visibleThemeIcon(page: Page, title: 'Switch to light mode' | 'Switch to dark mode') {
    return page.locator(`header span[title="${title}"]:visible`)
}

test('the demo dashboard still renders in the browser', async ({ page }) => {
    await page.goto('/dashboard-demo')

    await expect(page.locator('main a[aria-current="page"]')).toHaveText('Overview')
    await expect(page.getByRole('table', { name: 'HTTP APIs' })).toBeVisible()
})
