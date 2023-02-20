import { useFetch } from './fetch'

interface AppInfo {
    version: string
    activeServices: string[]
}

interface AppInfoResponse {
    data: AppInfo
    isLoading: Boolean
    error: string
    close: () => void
}

export function useAppInfo() : AppInfoResponse {
    return useFetch('/api/info')
}