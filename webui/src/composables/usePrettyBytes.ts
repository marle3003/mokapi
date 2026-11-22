interface FormatFunc {
    (value: unknown): string
}

export function usePrettyBytes() : {format: FormatFunc} {
    function format(value: unknown): string {
        // jacked from: https://github.com/sindresorhus/pretty-bytes
        if (typeof value !== 'number' || isNaN(value)) {
            return '0'
        }

        const neg = value < 0
        const units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

        let magnitude = neg ? -value : value

        if (magnitude < 1) {
            return (neg ? '-' : '') + magnitude + ' B'
        }

        const exponent = Math.min(
            Math.floor(Math.log(magnitude) / Math.log(1000)),
            units.length - 1
        )
        magnitude = magnitude / Math.pow(1000, exponent)
        const amount = Math.round(magnitude) !== magnitude ? magnitude.toFixed(2) : magnitude
        const unit = units[exponent]

        return (neg ? '-' : '') + amount + ' ' + unit
    }

    return { format }
}