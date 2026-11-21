import { getCurrentInstance, ref } from 'vue'

// Replacement for '@/router' in component tests: the real router imports most views. Use it as
//   vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// (the factory can't be shared as a value, because vi.mock is hoisted). The test mounts with a
// real memory router, see mountAt in mail-dashboard-api.ts.
//
// The route comes from the app (`$route`) and not from an import of vue-router: vi.resetModules
// gives every test its own vue-router, but the factory runs only once.
export function routerMock() {
  return {
    // refresh-manager.ts reads the refresh interval of the current route
    default: { currentRoute: ref({ query: {} }) },
    useRoute: () => {
      const route = getCurrentInstance()!.appContext.config.globalProperties.$route
      return {
        get params() { return route.params },
        service: (name: string, type: string) => ({ name: `${type}Service`, params: { service: name } }),
      }
    },
  }
}
