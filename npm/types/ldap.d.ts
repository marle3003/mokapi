type LdapEventHandler = (request: LdapSearchRequest, response: LdapSearchResponse) => boolean

declare module 'mokapi/ldap' {
    enum SearchScope {
        BaseObject,
        SingleLevel,
        WholeSubtree
    }
    enum ResultCode {
        Success = 0,
        OperationsError = 1,
        ProtocolError = 2,
        SizeLimitExceeded = 3,
        AuthMethodNotSupported = 4,
        CannotCancel = 121
    }
}

declare interface LdapSearchRequest {
    baseDN: string
    scope: LdapSearchScope,
    dereferencePolicy: number,
    sizeLimit: number,
    timeLimit: number,
    typesOnly: number,
    filter: string
    attributes: string[]
}

declare interface LdapSearchResponse{
    results: LdapSearchResult[]
    status: LdapResultStatus
    message: string
}

declare interface LdapSearchResult {
    dn: string
    attributes: { [name: string]: string[] }
}

declare enum LdapSearchScope {
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