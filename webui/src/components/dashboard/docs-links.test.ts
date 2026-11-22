import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { mount } from '@vue/test-utils'

// The dashboard can be hosted under a path prefix (<base href>). A root-absolute href="/docs/..."
// ignores <base>, so docs links must go through sitePath().
const prefix = '/mokapi/dashboard'

vi.mock('@/composables/dashboard', () => ({
  useDashboard: () => ({
    dashboard: ref({
      getJobs: () => ({ data: ref([]), isLoading: ref(false), close: vi.fn() }),
      getConfig: () => ({
        config: ref({ id: 'c1', url: 'file://a.yml', provider: 'file', time: '2024-01-01T10:00:00Z', tags: [] }),
        isLoading: ref(false),
        close: vi.fn(),
      }),
      getConfigData: () => ({ data: ref(''), isLoading: ref(false), filename: ref('a.yml'), close: vi.fn() }),
    }),
  }),
  getRouteName: (name: string) => computed(() => name),
}))
vi.mock('vue-router', () => ({ useRoute: () => ({ params: { id: 'c1' } }) }))
vi.mock('@/router', () => ({ useRoute: () => ({ meta: { mode: 'demo' } }) }))
vi.mock('@/components/dashboard/SourceView.vue', () => ({ default: { template: '<div />' } }))

import Actions from './Actions.vue'
import Config from './Config.vue'
import JobOverview from './job/JobOverview.vue'
import JobTags from './job/blocks/JobTags.vue'
import HttpExport from './http/export/OpenAPI.vue'
import KafkaExport from './kafka/export/AsyncAPI.vue'
import MqttExport from './mqtt/export/AsyncAPI.vue'
import WebsocketExport from './websocket/export/AsyncAPI.vue'

const stubs = { RouterLink: { template: '<a><slot /></a>' } }
const action = { tags: { event: 'custom' }, parameters: [], duration: 1, logs: [] } as unknown as Action
const exportProps = { serviceName: 'Petstore' }

const components = [
  ['Actions', Actions, { actions: [action] }],
  ['Config', Config, {}],
  ['JobOverview', JobOverview, {}],
  ['JobTags', JobTags, { tags: {} }],
  ['http export', HttpExport, exportProps],
  ['kafka export', KafkaExport, exportProps],
  ['mqtt export', MqttExport, exportProps],
  ['websocket export', WebsocketExport, exportProps],
] as const

describe('docs links of the dashboard below a path prefix', () => {
  beforeEach(() => {
    document.head.innerHTML = `<base href="${prefix}/">`
  })
  afterEach(() => {
    document.head.innerHTML = ''
  })

  it.each(components)('%s', (_name, component, props) => {
    const wrapper = mount(component, { props: props as never, global: { stubs } })

    const docsLinks = wrapper.findAll('a').map((a) => a.attributes('href') ?? '').filter((href) => href.includes('/docs/'))

    expect(docsLinks.length).toBeGreaterThan(0)
    for (const href of docsLinks) {
      expect(href).toMatch(new RegExp(`^${prefix}/docs/`))
    }
  })
})
