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
	scope: SearchScope
	dereferencePolicy: number
	sizeLimit: number
	timeLimit: number
	typesOnly: boolean
	filter: string
	attributes: string[]
}

declare interface LdapSearchResponse {
    results: LdapSearchResult[]
    status: number
    message: string
}

declare interface LdapSearchResult {
    dn: string
    attributes: { [name: string]: string }
}

declare enum SearchScope {
    BaseObject,
    SingleLevel,
    WholeSubtree
}

declare enum LdapResultStatus {
    Success = 0,
    OperationsError = 1,
    ProtocolError = 2,
    SizeLimitExceeded = 3,
    AuthMethodNotSupported = 4,
    CannotCancel = 121
}