// The route a search result opens, `undefined` for a result without detail page (HTTP log entry,
// unknown type or event namespace). Used by the cards of the search view and the rows of the panel.
export function resultRoute(item: SearchItem) {
    const params = item.params
    switch (params.type) {
        case 'http': return httpRoute(params)
        case 'kafka': return params.topic
            ? { name: 'kafkaTopic', params }
            : { name: 'kafkaService', params }
        case 'mail': return params.mailbox
            ? { name: 'smtpMailbox', params: { name: params.mailbox, ...params } }
            : { name: 'mailService', params }
        case 'ldap': return { name: 'ldapService', params }
        case 'config': return { name: 'config', params }
        case 'event': return eventRoute(params)
        default: return undefined
    }
}

function httpRoute(params: SearchItem['params']) {
    if (!params.path) {
        return { name: 'httpService', params }
    }
    const endpoint = params.path.split('/')
    endpoint.shift() // path starts with a slash: remove first empty entry
    if (params.method) {
        endpoint.push(params.method.toLowerCase())
    }
    return { name: 'httpEndpoint', params: { endpoint, ...params } }
}

function eventRoute(params: SearchItem['params']) {
    switch (params['traits.namespace']) {
        case 'http': return params['traits.type'] === 'log' ? undefined : { name: 'httpRequest', params }
        case 'kafka': return { name: 'kafkaMessage', params }
        case 'mail': return { name: 'smtpMail', params: { ...params, id: params.messageId } }
        case 'ldap': return { name: 'ldapRequest', params }
        default: return undefined
    }
}
