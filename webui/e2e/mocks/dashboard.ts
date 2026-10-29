import { app, sleep } from 'mokapi'
import { get, post, fetch } from 'mokapi/http'
import { clusters as kafkaClusters, events as kafkaEvents, configs as kafkaConfigs } from './kafka'
import { apps as httpServices, events as httpEvents, configs as httpConfigs } from './http'
import { services as mailServices, mailEvents } from './mail'
import { servers as ldapServers, searches } from './ldap'
import { clusters as mqttClusters, events as mqttEvents } from './mqtt'
import { services as websocketServices, events as websocketEvents } from './websocket'
import { sum, max } from './metrics'
import { getInfo } from './helpers'
import { jobs, config as jobConfig } from './jobs'
import { match, type Query } from '../../lib/events'

export default async function() {
    const apiBaseUrl = `http://localhost:8091/mokapi`
    const { version, buildTime } = get(`${apiBaseUrl}/api/info`).json()

    app.http()
        .use((_, res) => {
            res.headers["Access-Control-Allow-Methods"] = "*"
            res.headers["Access-Control-Allow-Headers"] = "*"
            res.headers["Access-Control-Allow-Origin"] = "*"
            res.headers['Mokapi-Version'] = version
            res.headers['Mokapi-Build-Time'] = buildTime
        })
        .operation('info', (_, res) => {
            res.data = { 
                version: "0.11.0", 
                activeServices: ["http", "kafka", "ldap", "mail", "mqtt", "websocket"],
                search: { enabled: true }
            }
        })
        .operation('services', (req, res) => {
            res.data = getServices(req.query['type'])
        })
        .operation('events', (req, res) => {
            const page = req.query['page'] ? parseInt(req.query['page']) : 1
            const size = req.query['size'] ? parseInt(req.query['size']) : 25

            const items = events.filter(x => match(x, query(req.query)))
            res.data = {
                items: items.slice((page - 1) * size, page * size),
                page: page,
                size: size,
                total: items.length
            }
        })
        .operation('event', (req, res) => {
            const event = events.find(x => x.id === req.path['id'])
            if (!event){
                res.rebuild(404)
                return
            }
            res.data = event
        })
        .operation('example', (req, res) => {
            const resp = post(`${apiBaseUrl}/api/schema/example`, req.body)
            res.statusCode = resp.statusCode
            res.headers = resp.headers
            res.body = resp.body
        })
        .operation('validate', (req, res) => {
            sleep(1000)
            let url = `${apiBaseUrl}/api/schema/validate`
            if (req.query['outputFormat']) {
                url += '?outputFormat=' + req.query['outputFormat']
            }
            const resp = post(url, req.body)
            res.statusCode = resp.statusCode
            res.headers = resp.headers
            res.body = resp.body
            res.data = null
        })
        .operation('configs', (req, res) => {
            res.data = configs
        })
        .operation('config', (req, res) => {
            const config = configs.find(x => x.id === req.path.id)
            if (!config) {
                res.rebuild(404)
                return
            }
            res.data = config
        })
        .operation('configData', (req, res) => {
            const config = configData[req.path.id]
            if (!config) {
                res.rebuild(404)
                return
            }
            res.data = config.data
            res.headers = config.headers
        })

    for (const config of configs) {
        if (!config.data || !config.data.startsWith('http://localhost:8090')) {
            continue
        }
        const r = await fetch(config.data)
        configData[config.id] = {
            data:   r.body,
            headers: {
                'Content-Disposition': `inline; filename="${config.filename}"`
            }
        }
    }
}

const configs = [
    ...Object.values(httpConfigs),
    ...Object.values(kafkaConfigs),
    ...Object.values(jobConfig)
]

const configData: Record<string, { data: string, headers: Record<string, string> }> = {}


function getServices(type: string) {
    const services = []
    if (!type || type === 'http') {
        services.push(...httpServices.map(x => {
            return {
                ...getInfo(x, 'http'),
                metrics: {
                    requests_total: sum('http_requests_total', { name: 'service', value: x.name }),
                    requests_errors_total: sum('http_requests_errors_total', { name: 'service', value: x.name }),
                    request_timestamp: max('http_request_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'kafka') {
        services.push(...kafkaClusters.map(x => {
            return {
                ...getInfo(x, 'kafka'),
                metrics: {
                    messages_total: sum('kafka_messages_total', { name: 'service', value: x.name }),
                    message_timestamp: max('kafka_message_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'mail') {
        services.push(...mailServices.map(x => {
            return {
                ...getInfo(x, 'mail'),
                metrics: {
                    mails_total: sum('mail_mails_total', { name: 'service', value: x.name }),
                    mail_timestamp: max('mail_mail_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'ldap') {
        services.push(...ldapServers.map(x => {
            return {
                ...getInfo(x, 'ldap'),
                metrics: {
                    requests_total: sum('ldap_requests_total', { name: 'service', value: x.name }),
                    request_timestamp: max('ldap_request_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'mqtt') {
        services.push(...mqttClusters.map(x => {
            return {
                ...getInfo(x, 'mqtt'),
                metrics: {
                    messages_total: sum('mqtt_messages_total', { name: 'service', value: x.name }),
                    message_timestamp: max('mqtt_message_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'websocket') {
        services.push(...websocketServices.map(x => {
            return {
                ...getInfo(x, 'websocket'),
                metrics: {
                    messages_total: sum('websocket_messages_total', { name: 'service', value: x.name }),
                    message_timestamp: max('websocket_message_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }

    return services
}

const events = [
    ...httpEvents,
    ...kafkaEvents,
    ...mqttEvents,
    ...websocketEvents,
    ...mailEvents,
    ...searches,
    ...jobs
]

function query(q: Record<string, any>): Query {
    const result: Query = { labels: Object.keys(q.traits).map(x => ({ name: x, value: q.traits[x] })) }

    if (q.http) {
        result.http = {
            url: q.http.url,
            statusCode: q.http.statusCode,
            clientIP: q.http.clientIP
        }

        if (q.http.method) {
            result.http.methods = q.http.method.split(',')
        }
        if (q.http.requestHeader) {
            result.http.requestHeaders = []
            for (const h of q.http.requestHeader) {
                const kv = h.split(':', 2)
                result.http.requestHeaders.push({ name: kv[0], value: kv.length == 1 ? '' : kv[1] })
            }
        }
        if (q.http.responseHeader) {
            result.http.responseHeaders = []
            for (const h of q.http.responseHeader) {
                const kv = h.split(':', 2)
                result.http.responseHeaders.push({ name: kv[0], value: kv.length == 1 ? '' : kv[1] })
            }
        }
    }

    return result
}