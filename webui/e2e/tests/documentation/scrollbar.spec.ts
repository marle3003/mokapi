import type { Page } from "@playwright/test"
import { expect, test } from "../models/fixture-website"

// Headless Chromium starts with --hide-scrollbars, which gives every scrollbar a width of 0.
test.use({ launchOptions: { ignoreDefaultArgs: ['--hide-scrollbars'] } })

const dark = { thumb: 'rgba(255, 255, 255, 0.16)', thumbHover: 'rgba(255, 255, 255, 0.32)' }
const light = { thumb: 'rgba(0, 0, 0, 0.18)', thumbHover: 'rgba(0, 0, 0, 0.38)' }

const scrollbarThumb = (element: Element) => getComputedStyle(element).getPropertyValue('--scrollbar-thumb').trim()
const scrollbarThumbHover = (element: Element) => getComputedStyle(element).getPropertyValue('--scrollbar-thumb-hover').trim()

async function useLightTheme(page: Page) {
    await page.addInitScript(() => localStorage.setItem('theme', 'light'))
}

test.describe('Scrollbars', () => {

    test.describe('colors', () => {
        test('dark theme uses a muted white thumb, stronger on hover', async ({ page }) => {
            await page.goto('/docs/welcome')

            const root = page.locator('html')
            await expect(root).toHaveAttribute('data-theme', 'dark')
            expect(await root.evaluate(scrollbarThumb)).toBe(dark.thumb)
            expect(await root.evaluate(scrollbarThumbHover)).toBe(dark.thumbHover)
        })

        test('light theme uses a muted black thumb, stronger on hover', async ({ page }) => {
            await useLightTheme(page)
            await page.goto('/docs/welcome')

            const root = page.locator('html')
            await expect(root).toHaveAttribute('data-theme', 'light')
            expect(await root.evaluate(scrollbarThumb)).toBe(light.thumb)
            expect(await root.evaluate(scrollbarThumbHover)).toBe(light.thumbHover)
        })

        test('code blocks keep the dark thumb in light theme, because they are dark in both themes', async ({ page }) => {
            await useLightTheme(page)
            await page.goto('/docs/get-started/installation')

            await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')
            const codeBlock = page.locator('pre').first()
            expect(await codeBlock.evaluate(scrollbarThumb)).toBe(dark.thumb)
            expect(await codeBlock.evaluate(scrollbarThumbHover)).toBe(dark.thumbHover)
        })
    })

    test.describe('visibility', () => {
        test('the docs sidebar thumb is invisible at rest and visible while the pointer is over the sidebar', async ({ page }) => {
            await page.setViewportSize({ width: 1280, height: 600 })
            await page.goto('/docs/welcome')

            const sidebar = page.locator('aside.sidebar')
            await expect(sidebar).toBeVisible()
            await expect.poll(() => sidebar.evaluate((el: HTMLElement) => el.scrollHeight > el.clientHeight),
                { message: 'docs sidebar must overflow to show a scrollbar' }).toBe(true)

            const box = (await sidebar.boundingBox())!
            const barWidth = await sidebar.evaluate((el: HTMLElement) => el.offsetWidth - el.clientWidth)
            const barRegion = { x: Math.floor(box.x + box.width - barWidth), y: Math.ceil(box.y), width: barWidth, height: Math.floor(box.height) }
            const contentPoint = { x: box.x + box.width / 2, y: box.y + box.height / 2 }
            const outside = { x: 5, y: 5 }
            const thumbColor = () => sidebar.evaluate(el => getComputedStyle(el).getPropertyValue('--scrollbar-thumb-current').trim())

            await page.mouse.move(outside.x, outside.y)
            await expect.poll(thumbColor).toBe('transparent')
            const atRest = await page.screenshot({ clip: barRegion })

            await page.mouse.move(contentPoint.x, contentPoint.y)
            await expect.poll(thumbColor).toBe(dark.thumb)
            await expect.poll(async () => (await page.screenshot({ clip: barRegion })).equals(atRest),
                { message: 'thumb must appear while hovering the sidebar' }).toBe(false)

            await page.mouse.move(outside.x, outside.y)
            await expect.poll(thumbColor).toBe('transparent')
            await expect.poll(async () => (await page.screenshot({ clip: barRegion })).equals(atRest),
                { message: 'thumb must disappear after leaving the sidebar' }).toBe(true)
        })
    })

    test.describe('width', () => {
        test('the docs sidebar scrollbar is 6px wide', async ({ page }) => {
            await page.setViewportSize({ width: 1280, height: 600 })
            await page.goto('/docs/welcome')

            const sidebar = page.locator('aside.sidebar')
            await expect(sidebar).toBeVisible()
            await expect.poll(() => sidebar.evaluate((el: HTMLElement) => el.scrollHeight > el.clientHeight),
                { message: 'docs sidebar must overflow to show a scrollbar' }).toBe(true)
            const barWidth = await sidebar.evaluate((el: HTMLElement) => el.offsetWidth - el.clientWidth)
            expect(barWidth).toBe(5)
        })

        test('the page scrollbar keeps the default width', async ({ page }) => {
            await page.setViewportSize({ width: 1280, height: 600 })
            // Reference: the width of a native page scrollbar on a plain page without our CSS.
            await page.setContent('<div style="height: 5000px"></div>')
            const defaultWidth = await pageScrollbarWidth(page)
            expect(defaultWidth).toBeGreaterThan(0)

            await page.goto('/docs/welcome')

            await expect(page.locator('aside.sidebar')).toBeVisible()
            await expect.poll(() => pageScrollbarWidth(page)).toBe(defaultWidth)
        })
    })
})

function pageScrollbarWidth(page: Page) {
    return page.evaluate(() => window.innerWidth - document.documentElement.clientWidth)
}
