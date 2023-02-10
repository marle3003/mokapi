import http from 'http-status-codes'

export function usePrettyHttp() {
    function format(statusCode: number): string {
        return `${statusCode} ${http.getStatusText(statusCode)}`
    }

    return {format}
}