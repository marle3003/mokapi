import { computed } from 'vue'
import { useDashboard, type Mode } from './dashboard'
import { useMetrics } from './metrics'
import type { AppInfo } from '@/types/dashboard'

// Jobs, Faker and Search are only available under these conditions,
// in the classic tabs as well as in the activity bar.
export function sectionVisibility(appInfo: AppInfo, hasJobs: boolean, mode: Mode) {
    return {
        jobs: hasJobs,
        faker: mode === 'live',
        search: appInfo.search.enabled && mode !== 'demo',
    }
}

export function useSectionVisibility(appInfo: () => AppInfo) {
    const { dashboard, getMode } = useDashboard()
    const { sum } = useMetrics()
    const response = dashboard.value.getMetrics('app')
    const hasJobs = computed(() => sum(response.data, 'app_job_run_total') > 0)

    return computed(() => sectionVisibility(appInfo(), hasJobs.value, getMode()))
}
