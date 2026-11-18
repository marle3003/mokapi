import type { Locator, Page } from '@playwright/test'
import {
  expandValidationError,
  fillFilter,
  filterWithChip,
  logOf,
  openFilter,
  openFirstClientRequest,
  openRow,
  openTab,
  selectJson,
  type LogNames,
} from './steps.ts'

export type Shot = {
  /** what `--only` matches (substring of name or file) */
  name: string
  /** target below docs/, e.g. http/dashboard-request-log.png */
  file: string
  url: string
  /** brings the page into the state to capture and waits until its data is visible */
  steps?: (page: Page) => Promise<void>
  clip?: (page: Page) => Promise<{ x: number; y: number; width: number; height: number }>
  viewport?: { width: number; height: number }
  /** 2 by default; 1 for images shown at their pixel size */
  deviceScaleFactor?: number
  /** used by the website, not by a docs page: `file` points out of docs/ and `--check` doesn't expect a reference */
  website?: boolean
}

// panels next to a log need more width than the default 1400 px
const wide = { width: 1600, height: 900 }

const dashboard = '/dashboard'

// ---- get-started/dashboard.md and running.md ----

const petstore = 'Swagger%20Petstore'
const petstoreUrl = `${dashboard}/http/services/${petstore}`

const httpLog: LogNames = { log: 'Recent Requests', panel: 'Request details' }
const jobLog: LogNames = { log: 'Recent Jobs', panel: 'Job details' }

const getStarted: Shot[] = [
  {
    name: 'overview',
    file: 'get-started/dashboard-overview-mock-api.png',
    url: dashboard,
    steps: async (page) => {
      await page.locator('.metric-grid').getByRole('link').first().waitFor()
    },
  },
  {
    name: 'search',
    file: 'get-started/dashboard-search.png',
    url: dashboard,
    viewport: { width: 1400, height: 1500 },
    steps: async (page) => {
      await page.getByRole('navigation', { name: 'Activity bar' }).getByRole('button', { name: 'Search' }).click()
      const panel = page.getByRole('complementary', { name: 'Search' })
      const input = panel.getByRole('textbox', { name: 'Search' })
      await input.fill('pet')
      await input.press('Enter')
      await panel.getByTestId('search-count').waitFor()
    },
  },
  {
    name: 'http-endpoints-by-tag',
    file: 'get-started/dashboard-http-endpoints-by-tag.png',
    url: petstoreUrl,
    viewport: { width: 1400, height: 1500 },
    steps: async (page) => {
      await openTab(page, 'Paths')
      await page.getByRole('button', { name: 'By tag' }).click()
      await page.getByRole('table', { name: 'Paths' }).locator('tbody tr').first().waitFor()
    },
  },
  {
    name: 'jobs-log-panel',
    file: 'get-started/dashboard-jobs-log-panel.png',
    url: `${dashboard}/jobs#tab-runs`,
    steps: async (page) => {
      await openRow(page, jobLog, (rows) => rows.filter({ hasText: 'kafka.ts' }), 3)
    },
  },
  {
    name: 'job-details',
    file: 'get-started/dashboard-job-details.png',
    url: `${dashboard}/jobs#tab-runs`,
    steps: async (page) => {
      const { panel } = await openRow(page, jobLog, (rows) => rows.filter({ hasText: 'kafka.ts' }), 3)
      await panel.getByRole('tab', { name: 'Details', exact: true }).click()
      await panel.getByText('Next Run').waitFor()
    },
    // the Tags table shows the path of the script on this machine
    clip: async () => ({ x: 0, y: 0, width: 1400, height: 630 }),
  },
  {
    name: 'endpoint-example-validate',
    file: 'get-started/dashboard-endpoint-example-validate.png',
    url: `${petstoreUrl}/pet/post`,
    viewport: { width: 1400, height: 1150 },
    steps: async (page) => {
      await selectJson(page)
    },
  },
  {
    name: 'validate-error',
    file: 'get-started/dashboard-validate-error.png',
    url: `${petstoreUrl}/pet/post`,
    steps: async (page) => {
      await selectJson(page)
      await page.getByRole('button', { name: 'Example & Validate' }).first().click()
      const dialog = page.getByRole('dialog', { name: 'Data Validator' })
      await dialog.getByRole('button', { name: 'Example', exact: true }).click()
      // a pet whose name is a number does not match the specification
      await dialog.locator('.ace_text-input').focus()
      await page.keyboard.press('ControlOrMeta+a')
      await page.keyboard.type('{"name": 42, "photoUrls": []')
      await dialog.getByRole('button', { name: 'Validate', exact: true }).click()
      await dialog.getByRole('alert').first().waitFor()
      await page.waitForTimeout(800) // the dialog fades in
    },
  },
  {
    name: 'petstore-endpoints',
    file: 'get-started/petstore-endpoints.png',
    url: petstoreUrl,
    viewport: { width: 1400, height: 1300 },
    steps: async (page) => {
      await openTab(page, 'Paths')
      await page.getByRole('table', { name: 'Paths' }).locator('tbody tr').first().waitFor()
    },
  },
  {
    name: 'petstore-request',
    file: 'get-started/petstore-request.png',
    url: `${dashboard}/http/log/${petstore}`,
    viewport: { width: 1800, height: 1000 },
    steps: async (page) => {
      const { panel } = await openRow(page, httpLog, (rows) => rows.filter({ has: page.getByText('http://localhost/v2/pet', { exact: true }) }), 3)
      await panel.getByRole('tab', { name: 'Request', exact: true }).click()
    },
  },
]

// ---- http/dashboard.md ----

const http: Shot[] = [
  {
    name: 'http-mock-rest-api',
    file: 'http/dashboard-mock-rest-api.png',
    url: petstoreUrl,
    viewport: { width: 1400, height: 1300 },
    steps: async (page) => {
      await page.getByRole('table').first().waitFor()
    },
  },
  {
    name: 'http-request-log',
    file: 'http/dashboard-request-log.png',
    url: petstoreUrl,
    steps: async (page) => {
      await openTab(page, 'Requests')
      await logOf(page, httpLog).rows.first().waitFor()
    },
  },
  {
    name: 'http-request-log-panel',
    file: 'http/dashboard-request-log-panel.png',
    url: `${dashboard}/http/log/${petstore}`,
    viewport: { width: 1800, height: 1000 },
    steps: async (page) => {
      // Mokapi keeps the body of a response only for an error
      await openRow(page, httpLog, (rows) => rows.filter({ hasText: '400 Bad Request' }), 3)
    },
  },
  {
    name: 'http-request-log-filter',
    file: 'http/dashboard-request-log-filter.png',
    url: `${dashboard}/http/log/${petstore}`,
    steps: async (page) => {
      await filterWithChip(page, httpLog, { field: 'URL', text: '/pet' }, 'Status Code')
    },
  },
  {
    name: 'http-endpoint-pinned',
    file: 'http/dashboard-endpoint-pinned.png',
    url: `${petstoreUrl}/pet/%7BpetId%7D/get`,
    steps: async (page) => {
      await page.getByRole('navigation', { name: 'Tabs' }).getByRole('button', { name: /^Pin/ }).click()
    },
  },
]

// ---- kafka, mqtt, websocket, mail and ldap: the same four views per protocol ----

const messageLog: LogNames = { log: 'Recent Messages', panel: 'Message details' }
const clientRequestLog: LogNames = { log: 'Recent Requests', panel: 'Request details' }

const kafkaService = 'Kafka%20Order%20Service%20API'
const kafka: Shot[] = [
  {
    name: 'kafka-message-log',
    file: 'kafka/dashboard-message-log.png',
    url: `${dashboard}/kafka/service/${kafkaService}`,
    steps: async (page) => {
      await openTab(page, 'Messages')
      await logOf(page, messageLog).rows.first().waitFor()
    },
  },
  {
    name: 'kafka-message-log-panel',
    file: 'kafka/dashboard-message-log-panel.png',
    url: `${dashboard}/kafka/log/${kafkaService}`,
    viewport: wide,
    steps: async (page) => {
      const { panel } = await openRow(page, messageLog, (rows) => rows, -1)
      await panel.getByRole('tab', { name: 'Value', exact: true }).click()
    },
  },
  {
    name: 'kafka-message-log-filter',
    file: 'kafka/dashboard-message-log-filter.png',
    url: `${dashboard}/kafka/log`,
    steps: async (page) => {
      await filterWithChip(page, messageLog, { field: 'Cluster' }, 'Partition')
    },
  },
  {
    name: 'kafka-client-requests',
    file: 'kafka/dashboard-client-requests.png',
    url: `${dashboard}/kafka/service/${kafkaService}`,
    viewport: wide,
    steps: async (page) => {
      await openFirstClientRequest(page, clientRequestLog, 'Request')
    },
  },
]

const mqttService = 'Smart%20Home%20MQTT%20API'
const mqttRequestLog: LogNames = { log: 'Connection & Subscription Requests', panel: 'Request details' }
const mqtt: Shot[] = [
  {
    name: 'mqtt-message-log',
    file: 'mqtt/dashboard-message-log.png',
    url: `${dashboard}/mqtt/service/${mqttService}`,
    steps: async (page) => {
      await openTab(page, 'Messages')
      await logOf(page, messageLog).rows.first().waitFor()
    },
  },
  {
    name: 'mqtt-message-log-panel',
    file: 'mqtt/dashboard-message-log-panel.png',
    url: `${dashboard}/mqtt/log/${mqttService}`,
    viewport: wide,
    steps: async (page) => {
      const { panel } = await openRow(page, messageLog, (rows) => rows, -1)
      await panel.getByRole('tab', { name: 'Value', exact: true }).click()
    },
  },
  {
    name: 'mqtt-message-log-filter',
    file: 'mqtt/dashboard-message-log-filter.png',
    url: `${dashboard}/mqtt/log`,
    steps: async (page) => {
      await filterWithChip(page, messageLog, { field: 'Cluster' }, 'Topic')
    },
  },
  {
    name: 'mqtt-client-requests',
    file: 'mqtt/dashboard-client-requests.png',
    url: `${dashboard}/mqtt/service/${mqttService}`,
    viewport: wide,
    steps: async (page) => {
      await openFirstClientRequest(page, mqttRequestLog, 'Request')
    },
  },
]

const websocketService = 'WebSocket%20Chat%20API'
const websocketRequestLog: LogNames = { log: 'Connection Requests', panel: 'Request details' }
const websocket: Shot[] = [
  {
    name: 'websocket-message-log',
    file: 'websocket/dashboard-message-log.png',
    url: `${dashboard}/websocket/service/${websocketService}`,
    steps: async (page) => {
      await openTab(page, 'Messages')
      await logOf(page, messageLog).rows.first().waitFor()
    },
  },
  {
    name: 'websocket-message-log-panel',
    file: 'websocket/dashboard-message-log-panel.png',
    url: `${dashboard}/websocket/log/${websocketService}`,
    viewport: wide,
    steps: async (page) => {
      const { panel } = await openRow(page, messageLog, (rows) => rows.filter({ hasText: 'Hello there!' }), -1)
      await panel.getByRole('tab', { name: 'Value', exact: true }).click()
    },
  },
  {
    name: 'websocket-message-validation',
    file: 'websocket/dashboard-message-validation.png',
    url: `${dashboard}/websocket/log/${websocketService}`,
    viewport: wide,
    steps: async (page) => {
      const { panel } = await openRow(page, messageLog, (rows) => rows.filter({ hasText: '"text":42' }), -1)
      await panel.getByRole('tab', { name: 'Validation', exact: true }).click()
      await expandValidationError(page, panel, 'incoming')
    },
  },
  {
    name: 'websocket-message-log-filter',
    file: 'websocket/dashboard-message-log-filter.png',
    url: `${dashboard}/websocket/log`,
    steps: async (page) => {
      await filterWithChip(page, messageLog, { field: 'Service' }, 'Path')
    },
  },
  {
    name: 'websocket-client-requests',
    file: 'websocket/dashboard-client-requests.png',
    url: `${dashboard}/websocket/service/${websocketService}`,
    viewport: wide,
    steps: async (page) => {
      await openFirstClientRequest(page, websocketRequestLog, '')
    },
    // the event handlers below show the path of the script on this machine
    clip: async () => ({ x: 0, y: 0, width: 1600, height: 760 }),
  },
]

const mailService = 'Mail%20Server'
const mailLog: LogNames = { log: 'Recent Mails', panel: 'Mail details' }
const mail: Shot[] = [
  {
    name: 'mail-log',
    file: 'mail/dashboard-mail-log.png',
    url: `${dashboard}/mail/service/${mailService}`,
    steps: async (page) => {
      await openTab(page, 'Mails')
      await logOf(page, mailLog).rows.first().waitFor()
    },
  },
  {
    name: 'mail-log-panel',
    file: 'mail/dashboard-mail-log-panel.png',
    url: `${dashboard}/mail/log/${mailService}`,
    viewport: { width: 2100, height: 1000 },
    steps: async (page) => {
      const { panel } = await openRow(page, mailLog, (rows) => rows.filter({ hasText: 'Check Out Our New Arrivals!' }), 4)
      await panel.getByRole('tab', { name: 'Body', exact: true }).click()
    },
  },
  {
    name: 'mail-rejected',
    file: 'mail/dashboard-mail-rejected.png',
    url: `${dashboard}/mail/log/${mailService}`,
    viewport: wide,
    steps: async (page) => {
      const rejected = (rows: Locator) => rows.filter({ has: page.getByTitle('Sender domain spam.example is blocked') })
      const { panel } = await openRow(page, mailLog, rejected, 4)
      await panel.getByRole('tab', { name: 'Details', exact: true }).click()
    },
  },
  {
    name: 'mail-log-filter',
    file: 'mail/dashboard-mail-log-filter.png',
    url: `${dashboard}/mail/log`,
    steps: async (page) => {
      await logOf(page, mailLog).rows.first().waitFor()
      await openFilter(page, mailLog, 'Service')
      await fillFilter(page, 'Service', '')
      await page.getByRole('list', { name: 'Active filters' }).getByRole('button').first().click()
    },
  },
]

const ldapService = 'HR%20Employee%20Directory'
const ldapLog: LogNames = { log: 'Recent Requests', panel: 'Request details' }
const ldap: Shot[] = [
  {
    name: 'ldap-request-log',
    file: 'ldap/dashboard-request-log.png',
    url: `${dashboard}/ldap/service/${ldapService}`,
    viewport: { width: 1600, height: 1100 },
    steps: async (page) => {
      await openTab(page, 'Requests')
      await logOf(page, ldapLog).rows.first().waitFor()
    },
  },
  {
    name: 'ldap-request-log-panel',
    file: 'ldap/dashboard-request-log-panel.png',
    url: `${dashboard}/ldap/log/${ldapService}`,
    viewport: wide,
    steps: async (page) => {
      const bind = (rows: Locator) => rows.filter({ has: page.getByText('Bind', { exact: true }) })
      const { panel } = await openRow(page, ldapLog, bind, 4)
      await panel.getByRole('tab', { name: 'Request', exact: true }).click()
    },
  },
  {
    name: 'ldap-search-response',
    file: 'ldap/dashboard-search-response.png',
    url: `${dashboard}/ldap/log/${ldapService}`,
    viewport: { width: 2000, height: 900 },
    steps: async (page) => {
      const { panel } = await openRow(page, ldapLog, (rows) => rows.filter({ hasText: 'Search' }), 4)
      await panel.getByRole('tab', { name: 'Response', exact: true }).click()
      await panel.getByRole('table').getByRole('row').nth(1).getByText('9', { exact: true }).click()
      await page.getByRole('dialog').waitFor()
      await page.waitForTimeout(800) // the dialog fades in
    },
  },
  {
    name: 'ldap-request-log-filter',
    file: 'ldap/dashboard-request-log-filter.png',
    url: `${dashboard}/ldap/log`,
    viewport: wide,
    steps: async (page) => {
      await filterWithChip(page, ldapLog, { field: 'Service' }, 'Operation')
    },
  },
]

// ---- faker explorer: get-started/dashboard.md, test-data.md and the home page ----

// customerEmail resolves to the email node (customerEmailAddress does too), sku to the custom node
// of demo-configs/faker.ts
const orderLineSchema = (...properties: string[]) =>
  JSON.stringify({ type: 'object', required: properties, properties: Object.fromEntries(properties.map((name) => [name, { type: 'string' }])) })

async function generateOrderLine(page: Page, schema: string) {
  const playground = page.getByRole('region', { name: 'Playground' })
  await playground.getByLabel('Schema').fill(schema)
  await playground.getByRole('button', { name: 'Generate' }).click()
  await playground.getByRole('table', { name: 'Matches' }).locator('tbody tr').first().waitFor()
}

const faker: Shot[] = [
  {
    name: 'faker-explorer',
    file: 'get-started/dashboard-faker-explorer.png',
    url: `${dashboard}/faker?node=/sku`,
    viewport: { width: 1600, height: 1000 },
    steps: async (page) => {
      await page.getByRole('region', { name: 'Node details' }).getByText('/sku', { exact: true }).waitFor()
      await generateOrderLine(page, orderLineSchema('customerEmail', 'firstName', 'city', 'sku'))
    },
  },
  {
    name: 'faker-home',
    file: '../webui/public/website/mock-realistic-test-data.png',
    url: `${dashboard}/faker?node=/sku`,
    // the list scrolls to the selected node at its bottom, which has to lie inside the clip
    viewport: { width: 1600, height: 720 },
    clip: async () => ({ x: 350, y: 115, width: 1200, height: 583 }),
    deviceScaleFactor: 1,
    website: true,
    steps: async (page) => {
      await page.getByRole('region', { name: 'Node details' }).getByText('/sku', { exact: true }).waitFor()
      await generateOrderLine(page, orderLineSchema('customerEmail', 'firstName', 'sku'))
    },
  },
]

export const shots: Shot[] = [...getStarted, ...faker, ...http, ...kafka, ...mqtt, ...websocket, ...mail, ...ldap]
