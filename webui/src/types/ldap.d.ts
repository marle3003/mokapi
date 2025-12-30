declare interface LdapService extends Service {
    server: string
}

declare interface LdapEventData {
    request: LdapBindRequest | LdapUnbindRequest | LdapSearchRequest | LdapModifyRequest | LdapAddRequest | LdapDeleteRequest | LdapModifDNRequest | LdapCompareRequest
    response: LdapSearchResponse | LdapResponse
    duration: number
    actions: Action[]
}

declare interface LdapBindRequest {
    operation: 'Bind'
    version: string
    name: string
    password: string
    auth: string
}

declare interface LdapUnbindRequest {
    operation: 'Unbind'
}

declare interface LdapSearchRequest {
    operation: 'Search'
    baseDN: string
	scope: string
	dereferencePolicy: number
	sizeLimit: number
	timeLimit: number
	typesOnly: boolean
	filter: string
	attributes: string[]
}

declare interface LdapSearchResponse {
    results: LdapSearchResult[]
    status: string
}

declare interface LdapSearchResult {
    dn: string
    attributes: { [name: string]: string[] }
}

declare interface LdapModifyRequest {
    operation: 'Modify'
    dn: string
    items: LdapModifyItem[]
}

declare interface LdapModifyItem {
    modification: string
    attribute: LdapAttribute
}

declare interface LdapResponse {
    status: string
    matchedDn: string
    message: string
}

declare interface LdapAttribute {
    type: string
    values: string[]
}

declare interface LdapAddRequest {
    operation: 'Add'
    dn: string
    attributes: LdapAttribute[]
}

declare interface LdapDeleteRequest {
    operation: 'Delete'
    dn: string
}

declare interface LdapModifDNRequest {
    operation: 'ModifyDN'
	dn: string
	newRdn: string
	deleteOldDn: boolean
	newSuperiorDn: string
}

declare interface LdapCompareRequest {
    operation: 'Compare'
	dn: string
	attribute: string
    value: string
}