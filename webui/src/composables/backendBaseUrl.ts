export function useBackendBaseUrl() : string {
    let baseUrl = import.meta.env.VITE_BACKEND_BASE_URL
    if (baseUrl == '' && window.location.pathname != ''){
        const host = window.location.origin
        const root = window.location.pathname
        return (host + root).replace(/\/$/, '')
    }
    return baseUrl
}