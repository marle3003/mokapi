import http from 'http-status-codes'

export function usePrettyHttp() {
    function formatStatusCode(statusCode: number): string {
        return `${statusCode} ${http.getStatusText(statusCode)}`
    }
    function getClassByStatusCode(statusCode: number) {
        switch (true) {
            case statusCode >= 200 && statusCode < 300:
                return 'success'
            case statusCode >= 300 && statusCode < 400:
                return 'redirect'
            case statusCode >= 400 && statusCode < 500:
                return 'client-error'
            case statusCode >= 500 && statusCode < 600:
                return 'server-error'
        }
        return ''
    }

    return {formatStatusCode, getClassByStatusCode}
}