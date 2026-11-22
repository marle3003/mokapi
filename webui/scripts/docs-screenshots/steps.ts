import type { Locator, Page } from '@playwright/test'

// The steps use plain Playwright locators (roles and accessible names). The page objects in
// e2e/tests/models import each other without file extension, which Node doesn't resolve.

export type LogNames = { log: string; panel: string }

export function logOf(page: Page, { log, panel }: LogNames) {
  const region = page.getByRole('region', { name: log })
  return {
    region,
    rows: region.getByRole('table', { name: log }).locator('tbody tr'),
    panel: page.getByRole('region', { name: panel }),
    addFilter: region.getByRole('button', { name: 'Add filter' }),
  }
}

// A click on a cell without a link selects the row and opens its detail panel.
export async function openRow(page: Page, names: LogNames, row: (rows: Locator) => Locator, cell: number) {
  const view = logOf(page, names)
  const selected = row(view.rows)
  await selected.first().waitFor()
  await selected.first().getByRole('cell').nth(cell).click()
  await view.panel.waitFor()
  await page.waitForLoadState('networkidle') // the panel loads the details of the selected row
  return view
}

export async function openTab(page: Page, name: string) {
  await page.getByRole('tab', { name, exact: true }).click()
}

// Opens the popover of a field through "+ Add filter". The menu sometimes closes before the click
// arrives while the page still loads, so the steps run again a few times.
export async function openFilter(page: Page, names: LogNames, field: string) {
  const popover = page.getByRole('dialog', { name: field, exact: true })
  for (let attempt = 1; ; attempt++) {
    await logOf(page, names).addFilter.click()
    await page.getByRole('menu').getByRole('menuitem', { name: field, exact: true }).click()
    try {
      await popover.waitFor({ timeout: 5000 })
      return
    } catch (err) {
      if (attempt === 3) throw err
      await page.keyboard.press('Escape')
    }
  }
}

// Fills the popover of a filter: a select takes its last option, an input the given text.
export async function fillFilter(page: Page, field: string, text: string) {
  const popover = page.getByRole('dialog', { name: field, exact: true })
  const control = popover.getByRole('combobox').or(popover.getByRole('textbox')).first()
  if ((await control.evaluate((element) => element.tagName)) === 'SELECT') {
    await control.selectOption({ index: (await control.locator('option').count()) - 1 })
  } else {
    await control.fill(text)
  }
  await page.keyboard.press('Escape')
}

// Chips with data in the table and the popover of a second field open.
export async function filterWithChip(page: Page, names: LogNames, chip: { field: string; text?: string }, open: string) {
  await logOf(page, names).rows.first().waitFor()
  await openFilter(page, names, chip.field)
  await fillFilter(page, chip.field, chip.text ?? '')
  await openFilter(page, names, open)
}

// The client view of the first client of a service: its request log with the first request selected.
export async function openFirstClientRequest(page: Page, names: LogNames, tab: string) {
  await openTab(page, 'Clients')
  await page.getByRole('table', { name: 'Clients' }).locator('tbody tr').first().getByRole('link').first().click()
  const { panel } = await openRow(page, names, (rows) => rows, -1)
  if (tab) await panel.getByRole('tab', { name: tab, exact: true }).click()
}

// The order of the content types of a request body is random, the shots show JSON.
export async function selectJson(page: Page) {
  await page.getByRole('combobox').first().selectOption('application/json')
}

// The validation errors of a WebSocket message are collapsed, a click shows the error text.
export async function expandValidationError(page: Page, panel: Locator, messageId: string) {
  const entry = panel.getByRole('table', { name: 'Validation Errors' }).getByRole('row').filter({ hasText: messageId })
  await entry.click()
  await page.getByText('invalid type, expected string but got number').first().waitFor()
  await page.waitForTimeout(600) // the entry unfolds
}
