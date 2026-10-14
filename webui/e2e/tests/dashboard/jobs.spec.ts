import { useDashboard } from '../components/dashboard'
import { test, expect } from '../models/fixture-dashboard'

test.describe('Visit Jobs dashboard', () => {
    test.use({ colorScheme: 'dark' })

     test('Verifiy Job table', async ({ page }) => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.jobs.click()

        const table = page.getByRole('table', { name: 'Recent Jobs' });
        let rows = table.locator('> tbody > tr:not(.collapse-row)');
        await expect(rows).toHaveCount(3)
     })
})