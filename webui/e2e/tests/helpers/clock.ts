import type { Page } from '@playwright/test'

// Stops the installed clock shortly after "now", so only page.clock.runFor() lets timers fire.
export async function pauseClock(page: Page) {
    const now = await page.evaluate(() => Date.now())
    await page.clock.pauseAt(now + 100)
}
