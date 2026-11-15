import { endpointMethod } from '@/composables/dashboard-tabs'

type Params = Record<string, unknown>

// Pages that show the same for every visit, or whose id would not tell the reader anything.
const pageLabels: Record<string, string> = {
    dashboard: 'Overview',
    serviceList: 'Services',
    search: 'Search',
    http: 'HTTP',
    kafka: 'Kafka',
    ldap: 'LDAP',
    mail: 'Mail',
    mqtt: 'MQTT',
    websocket: 'Websocket',
    jobs: 'Jobs',
    configs: 'Configs',
    config: 'Config',
    tree: 'Faker',
    httpRequest: 'HTTP Request',
    kafkaRequest: 'Kafka Request',
    kafkaMessage: 'Kafka Message',
    ldapRequest: 'LDAP Request',
    smtpMail: 'Mail',
    mqttRequest: 'MQTT Request',
    mqttMessage: 'MQTT Message',
    websocketMessage: 'Websocket Message',
    websocketEvent: 'Websocket Event',
}

// Pages that are named after one of their route params.
const nameParams: Record<string, string> = {
    httpService: 'service',
    kafkaService: 'service',
    kafkaServer: 'server',
    kafkaTopic: 'topic',
    kafkaGroup: 'group',
    kafkaGroupMember: 'member',
    kafkaClient: 'clientId',
    ldapService: 'service',
    mailService: 'service',
    smtpMailbox: 'name',
    mqttService: 'service',
    mqttServer: 'server',
    mqttTopic: 'topic',
    mqttClient: 'clientId',
    websocketService: 'service',
    websocketServer: 'server',
    websocketChannel: 'channel',
    websocketClient: 'id',
}

// Short name of a dashboard page for tabs and pinned entries; the full title is their tooltip.
// Undefined when the route is not a dashboard page.
export function routeLabel(routeName: string | symbol | undefined | null, params: Params): string | undefined {
    if (typeof routeName !== 'string') {
        return undefined
    }
    const name = routeName.replace(/-demo$/, '')
    if (name === 'httpEndpoint') {
        return endpointPath(params.endpoint as string[] | undefined)
    }
    const nameParam = nameParams[name]
    return nameParam ? (params[nameParam] as string | undefined) : pageLabels[name]
}

// The operation's method is shown as a badge next to the label, not in it.
function endpointPath(endpoint: string[] | undefined): string {
    const segments = endpoint ?? []
    return '/' + (endpointMethod(segments) ? segments.slice(0, -1) : segments).join('/')
}
