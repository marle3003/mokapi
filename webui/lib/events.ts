interface Event {
    id: string;
    traits: Traits;
    time: string;
    data: any
}

interface Label {
    name: string
    value: string
}

export interface Query {
    labels: Label[]
    http?: HttpEventFilter
}

export interface HttpEventFilter {
    methods?: string[]
    url?: string
    requestHeaders?: { name: string, value: string }[]
    statusCode?: string
    responseHeaders?: { name: string, value: string }[]
    clientIP?: string
}

export function match(e: Event, q?: Query): boolean {
    if (!q) {
        return true
    }

    if (!matchTraits(e, q.labels)) {
        return false
    }
    if (q.http && e.traits.namespace === 'http') {
        console.log(q.http.methods, e)
        if (q.http.methods) {
            if (!q.http.methods.includes(e.data.request.method)) {
                return false
            }
        }
        if (q.http.url) {
            if (!e.data.request.url.match(q.http.url)) {
                return false
            }
        }
        if (q.http.requestHeaders) {
            if (!e.data.request.parameters) {
                return false
            }
            for (const h of q.http.requestHeaders) {
                const found = e.data.request.parameters.find((x: any) => x.type === 'header' && x.name === h.name)
                if (!found) {
                    return false
                }
                if (h.value !== '' && (!found.raw || !found.raw.match(h.value))) {
                    return false
                }
            }
        }
        if (q.http.responseHeaders) {
            if (!e.data.response.headers) {
                return false
            }
            for (const h of q.http.responseHeaders) {
                const found = e.data.response.headers[h.name]
                if (!found) {
                    return false
                }
                if (h.value !== '' && !found.match(h.value)) {
                    return false
                }
            }
        }
        if (q.http.statusCode) {
            const values = q.http.statusCode.split(',').map((x: string) => x.trim()) as string[]
            const statusCode = e.data.response.statusCode
            for (let val of values) {
                let not = false
                let match = false
                if (val[0] === '!') {
                    not = true
                    val = val.substring(1)
                }

                if (val.length === 3 && val.includes('xx')) {
                    val = `${val[0]}00-${val[0]}99`
                }

                if (val.includes('-')) {
                    const parts = val.split('-').map(x => x.trim())
                    if (parts.length != 2) {
                        continue
                    }
                    const min = parseInt(parts[0])
                    const max = parseInt(parts[1])
                    match = statusCode >= min && statusCode <= max
                } else {
                    match = statusCode.toString() === val
                }

                if (not === match) {
                    return false
                }
            }
        }
        if (q.http.clientIP) {
            const values = q.http.clientIP.split(',').map((x: string) => x.trim()) as string[]
            for (let val of values) {
                let not = false
                let match = false
                if (val[0] === '!') {
                    not = true
                    val = val.substring(1)
                }

                if (val.includes('/')) {
                    match = cidrMatch(e.data.clientIP, val)
                } else {
                    match = e.data.clientIP === val
                }

                if (not === match) {
                    return false
                }
            } 
        }
    }
    return true
}

export function matchTraits(e: ServiceEvent, labels: Label[]) {
    for (let label of labels) {
        if (e.traits[label.name] !== label.value) {
            return false
        }
    }
    return true
}

function cidrMatch(ip: string, cidr: string): boolean {
    const parts: string[] = cidr.split('/');
    if (parts.length !== 2) {
        return false
    }

    const [range, bitsStr] = parts;
    const bits: number = parseInt(bitsStr!, 10);

    if (isNaN(bits) || bits < 0 || bits > 32) {
        return false
    }

    const mask: number = ~(2 ** (32 - bits) - 1);

    return (ipToInt(ip) & mask) === (ipToInt(range!) & mask);
}

function ipToInt(ip: string): number {
    return ip
        .split('.')
        .map((octet) => parseInt(octet, 10))
        .reduce((acc, octet) => (acc << 8) + octet, 0);
}