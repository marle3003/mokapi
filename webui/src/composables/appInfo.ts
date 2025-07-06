import { useFetch } from './fetch'

interface AppInfo {
    version: string
    activeServices: string[]
    search: Search
}

export interface AppInfoResponse {
    data: AppInfo
    isLoading: boolean
    error: string
    close: () => void
}

export interface Search {
    enabled: boolean
}

export function useAppInfo() : AppInfoResponse {
    return useFetch('/api/info')
}