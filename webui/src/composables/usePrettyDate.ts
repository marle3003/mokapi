import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import duration from 'dayjs/plugin/duration'

dayjs.extend(relativeTime)
dayjs.extend(duration)

export function usePrettyDates() {
    function fromNow(value: number): string {
        return dayjs.unix(value).fromNow()
    }
    function format(value: number | string): string {
        if (typeof value == 'string') {
            return dayjs(value).format('YYYY-MM-DD HH:mm:ss')
        }
        return dayjs.unix(value).format('YYYY-MM-DD HH:mm:ss')
    }
    function duration(value: number): string{
        let d = dayjs.duration(value)
        if (d.seconds() < 1){
            return d.milliseconds() + ' [ms]'
        } else if (d.minutes() < 1){
            return d.seconds() + ' [sec]'
        }
        return d.minutes() + ' [min]'
    }

    return {fromNow, format, duration}
}