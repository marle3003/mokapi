import { useFetch } from './fetch'

interface AppInfo {
    version: string
    activeServices: string[]
}

interface AppInfoResponse {
    data: AppInfo
}

export function useAppInfo() : AppInfoResponse {
    return useFetch('/api/info')
}