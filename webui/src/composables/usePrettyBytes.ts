interface FormatFunc {
    (value: any): string
}

export function usePrettyBytes() : {format: FormatFunc} {
    function format(value: any): string {
        // jacked from: https://github.com/sindresorhus/pretty-bytes
        if (typeof value !== 'number' || isNaN(value)) {
            return '0'
        }

        let exponent
        let unit
        let neg = value < 0
        let units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

        if (neg) {
            value = -value
        }

        if (value < 1) {
            return (neg ? '-' : '') + value + ' B'
        }

        exponent = Math.min(
            Math.floor(Math.log(value) / Math.log(1000)),
            units.length - 1
        )
        value = (value / Math.pow(1000, exponent)).toFixed(2)
        unit = units[exponent]

        return (neg ? '-' : '') + value + ' ' + unit
    }

    return {format}
}