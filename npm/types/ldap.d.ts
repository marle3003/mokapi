type LdapEventHandler = (request: LdapSearchRequest, response: LdapSearchResponse) => boolean

declare interface LdapSearchRequest {
    baseDN: string
    scope: SearchScope,
    dereferencePolicy: number,
    sizeLimit: number,
    timeLimit: number,
    typesOnly: number,
    filter: string
    attributes: string[]
}

declare interface LdapSearchResponse{
    results: LdapSearchResult[]
    status: LdapStatus
    message: string
}

declare interface LdapSearchResult {
    dn: string
    attributes: { [name: string]: string[] }
}

declare enum SearchScope{
    BaseObject,
    SingleLevel,
    WholeSubtree
}

declare enum LdapStatus{
    Success,
    OperationsError,
}