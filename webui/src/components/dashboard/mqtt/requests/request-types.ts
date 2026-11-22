import type { Component } from 'vue'
import ConnectSummary from './ConnectSummary.vue'
import ConnectRequest from './ConnectRequest.vue'
import ConnectResponse from './ConnectResponse.vue'
import SubscribeSummary from './SubscribeSummary.vue'
import SubscribeRequest from './SubscribeRequest.vue'
import SubscribeResponse from './SubscribeResponse.vue'
import DisconnectSummary from './DisconnectSummary.vue'
import DisconnectRequest from './DisconnectRequest.vue'

// What the dashboard shows of an MQTT request by its packet type: the summary of a log row, and the
// request and response part of the details. DISCONNECT has no response.
export interface MqttRequestType {
    summary: Component
    request: Component
    response?: Component
}

export const mqttRequestTypes: { [packetType: number]: MqttRequestType } = {
    1: { summary: ConnectSummary, request: ConnectRequest, response: ConnectResponse },
    8: { summary: SubscribeSummary, request: SubscribeRequest, response: SubscribeResponse },
    14: { summary: DisconnectSummary, request: DisconnectRequest },
}
