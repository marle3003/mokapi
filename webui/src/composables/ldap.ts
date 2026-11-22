export type LdapTab = 'Request' | 'Response' | 'Details'
type LdapRequest = LdapEventData['request']

// The sub tabs of the panel of a request. Unbind and Delete have no request fields to show, only Search has a response to list.
export function ldapTabs(operation: LdapRequest['operation']): LdapTab[] {
    switch (operation) {
        case 'Unbind':
        case 'Delete':
            return ['Details']
        case 'Search':
            return ['Request', 'Response', 'Details']
        default:
            return ['Request', 'Details']
    }
}

// The sub tab belongs to the panel, not to the request: the next request opens on the same tab if it has it.
export function activeLdapTab(tabs: LdapTab[], selected: LdapTab): LdapTab {
    return tabs.includes(selected) ? selected : tabs[0]
}

// What a request is about at first sight: the filter of a search, the name of a bind, the DN of the rest. Unbind has none.
export function ldapHeadText(request: LdapRequest): string | undefined {
    switch (request.operation) {
        case 'Search':
            return request.filter
        case 'Bind':
            return request.name
        case 'Unbind':
            return undefined
        default:
            return request.dn
    }
}

export function useLdap() {
    
    function targetDn (data: LdapEventData) {
        switch (data.request.operation) {
            case 'Bind':
                return data.request.name;
            case 'Search':
                return data.request.baseDN
            case 'Modify':
            case 'Add':
            case 'Delete':
            case 'ModifyDN':
            case 'Compare':
                return data.request.dn
        }
    }

    function criteria(data: LdapEventData) {
        switch (data.request.operation) {
            case 'Search':
                return data.request.filter
            case 'Compare':
                return `${data.request.attribute} == ${data.request.value}`
            case 'Modify':
                const result = []
                for (const item of data.request.items) {
                    result.push(`${item.modification.toLocaleLowerCase()} ${item.attribute.type}`)
                }
                return result.join('<br />')
            case 'ModifyDN':
                return data.request.newRdn
        }
    }

    return { targetDn, criteria }
}