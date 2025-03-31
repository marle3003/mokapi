import http from 'http-status-codes'

export function usePrettyHttp() {
    function formatStatusCode(statusCode: string): string {
        if (statusCode === 'default') {
            return 'default'
        } else {
            return `${statusCode} ${http.getStatusText(statusCode)}`
        }
    }
    function getClassByStatusCode(statusCode: string) {
        if (typeof statusCode === 'string') {
            if (statusCode === 'default') {
                return 'default'
            }
        }

        const value = parseInt(statusCode)
        switch (true) {
            case value >= 200 && value < 300:
                return 'success'
            case value >= 300 && value < 400:
                return 'redirect'
            case value >= 400 && value < 500:
                return 'client-error'
            case value >= 500 && value < 600:
                return 'server-error'
            default:
                return ''
        }
    }

    return {formatStatusCode, getClassByStatusCode}
}