import { describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { mount } from '@vue/test-utils'

const job: JobInfo = {
  id: 'job-1',
  name: 'cleanup',
  schedule: 'every 5m',
  maxRuns: 0,
  runs: 2,
  state: 'scheduled',
  tags: { file: 'jobs.js', fileKey: 'abc' },
  lastRun: { eventId: 'event-1', time: '2024-01-01T10:00:00Z', duration: 10, status: 'success' },
}

vi.mock('@/composables/dashboard', () => ({
  useDashboard: () => ({ dashboard: ref({ getJobs: () => ({ data: ref([job]), isLoading: ref(false), close: vi.fn() }) }) }),
  getRouteName: (name: string) => computed(() => name),
}))

import JobOverview from './JobOverview.vue'

function mountOverview() {
  return mount(JobOverview, {
    global: { stubs: { RouterLink: { template: '<a class="router-link"><slot /></a>' } } },
  })
}

describe('JobOverview', () => {
  it('selects the job once when its name button is clicked', async () => {
    const overview = mountOverview()

    await overview.get('button.row-link').trigger('click')

    expect(overview.emitted('select')).toEqual([['cleanup']])
  })

  it('selects the job when another cell of the row is clicked', async () => {
    const overview = mountOverview()

    await overview.findAll('tbody td')[1].trigger('click')

    expect(overview.emitted('select')).toEqual([['cleanup']])
  })

  it('does not select the job when a link is clicked', async () => {
    const overview = mountOverview()

    for (const link of overview.findAll('a.router-link')) {
      await link.trigger('click')
    }

    expect(overview.emitted('select')).toBeUndefined()
  })
})
