import type { Component } from 'vue'
import BindRequest from './BindRequest.vue'
import SearchRequest from './SearchRequest.vue'
import SearchResponse from './SearchResponse.vue'
import ModifyRequest from './ModifyRequest.vue'
import AddRequest from './AddRequest.vue'
import ModifyDnRequest from './ModifyDnRequest.vue'
import CompareRequest from './CompareRequest.vue'

// What the dashboard shows of an LDAP request by its operation: the request and response blocks of the
// details. Unbind and Delete have no request block, only Search has a response block.
export interface LdapRequestType {
    request?: Component
    response?: Component
}

export const ldapRequestTypes: { [operation: string]: LdapRequestType } = {
    Bind: { request: BindRequest },
    Search: { request: SearchRequest, response: SearchResponse },
    Modify: { request: ModifyRequest },
    Add: { request: AddRequest },
    ModifyDN: { request: ModifyDnRequest },
    Compare: { request: CompareRequest },
}
