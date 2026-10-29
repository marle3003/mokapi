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
            default:
                undefined;
        }
    }

    return { targetDn, criteria }
}