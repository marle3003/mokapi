import dayjs from 'dayjs'

export const formatTimestamp = function(timestamp: number): string{
    return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

export const formatDateTime = function(s: string): string{
    return dayjs(s).format('YYYY-MM-DD HH:mm:ss')
}