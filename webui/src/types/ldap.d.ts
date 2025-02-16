declare interface LdapService extends Service {
    server: string
}

declare interface LdapEventData {
    request: LdapSearchRequest
    response: LdapSearchResponse
    duration: number
    actions: Action[]
}

declare interface LdapSearchRequest {
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
    message: string
}

declare interface LdapSearchResult {
    dn: string
    attributes: { [name: string]: string }
}