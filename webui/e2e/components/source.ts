import { test, Locator, expect } from "playwright/test"

export interface Source {
    lines?: ExpectedString
    size?: ExpectedString
    content: ExpectedString
    filename: ExpectedString
    clipboard: ExpectedString 
}

export function useSourceView(locator: Locator) {
    return {
        async test(expected: Source) {
            const source = locator.getByRole('region', { name: 'Source' })
            if (expected.lines) {
                await expect(source.getByLabel('Lines of Code')).toHaveText(expected.lines)
            } else {
                await expect(source.getByLabel('Lines of Code')).not.toBeVisible()
            }
            if (expected.size) {
                await expect(source.getByLabel('Size of Code')).toHaveText(expected.size)
            } else {
                await expect(source.getByLabel('Size of Code')).not.toBeVisible()
            }
            await expect(source.getByRole('region', { name: 'content' }).locator('.ace_content')).toHaveText(expected.content)

            await source.getByRole('button', { name: 'Copy raw content' }).click()
            let clipboardText = await locator.page().evaluate('navigator.clipboard.readText()')
            if (typeof expected.clipboard === 'string') {
                await expect(clipboardText).toContain(expected.clipboard)
            }else if (expected.clipboard instanceof RegExp) {
                await expect(clipboardText).toMatch(expected.clipboard)
            }

            await test.step('Check download', async () => {
                const [ download ] = await Promise.all([
                    locator.page().waitForEvent('download'),
                    source.getByRole('button', { name: 'Download raw content' }).click()
                ])
                await expect(download.suggestedFilename()).toBe(expected.filename)
            })
        }
    }
}