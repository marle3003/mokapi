import type { Component } from 'vue'
import ListOffsetsSummary from './ListOffsetsSummary.vue'
import ListOffsetsRequest from './ListOffsetsRequest.vue'
import ListOffsetsResponse from './ListOffsetsResponse.vue'
import FindCoordinatorSummary from './FindCoordinatorSummary.vue'
import FindCoordinatorRequest from './FindCoordinatorRequest.vue'
import FindCoordinatorResponse from './FindCoordinatorResponse.vue'
import JoinGroupSummary from './JoinGroupSummary.vue'
import JoinGroupRequest from './JoinGroupRequest.vue'
import JoinGroupResponse from './JoinGroupResponse.vue'
import SyncGroupSummary from './SyncGroupSummary.vue'
import SyncGroupRequest from './SyncGroupRequest.vue'
import SyncGroupResponse from './SyncGroupResponse.vue'
import InitProducerIdSummary from './InitProducerIdSummary.vue'
import InitProducerIdRequest from './InitProducerIdRequest.vue'
import InitProducerIdResponse from './InitProducerIdResponse.vue'

// What the dashboard shows of a Kafka request by its API key: the summary of a log row, and the
// request and response part of the details (props: request, response and the version of the
// request header, as far as the part needs them).
export interface KafkaRequestType {
    summary: Component
    request: Component
    response: Component
}

export const kafkaRequestTypes: { [apiKey: number]: KafkaRequestType } = {
    2: { summary: ListOffsetsSummary, request: ListOffsetsRequest, response: ListOffsetsResponse },
    10: { summary: FindCoordinatorSummary, request: FindCoordinatorRequest, response: FindCoordinatorResponse },
    11: { summary: JoinGroupSummary, request: JoinGroupRequest, response: JoinGroupResponse },
    14: { summary: SyncGroupSummary, request: SyncGroupRequest, response: SyncGroupResponse },
    22: { summary: InitProducerIdSummary, request: InitProducerIdRequest, response: InitProducerIdResponse },
}
