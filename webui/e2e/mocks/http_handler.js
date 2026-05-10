import { on, sleep } from 'mokapi'
import { clusters as kafkaClusters, events as kafkaEvents, configs as kafkaConfigs } from 'kafka.ts'
import { apps as httpServices, events as httpEvents, configs as httpConfigs } from 'services_http.js'
import { services as mailServices, mailEvents, getMail, getAttachment } from 'mail.js'
import { server as ldapServers, searches } from 'ldap.js'
import { clusters as mqttClusters, events as mqttEvents } from 'mqtt.js'
import { metrics, sum, max } from 'metrics.ts'
import { get, post, fetch } from 'mokapi/http'
import { base64 } from 'mokapi/encoding';

const configs = {}
const jobs = []

export default async function() {
    const apiBaseUrl = `http://localhost:8091/mokapi`
    const res = get(`${apiBaseUrl}/api/info`)
    const { version, buildTime } = res.json()

    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Methods"] = "*"
        response.headers["Access-Control-Allow-Headers"] = "*"
        response.headers["Access-Control-Allow-Origin"] = "*"
        response.headers['Mokapi-Version'] = version
        response.headers['Mokapi-Build-Time'] = buildTime

        switch (request.operationId) {
            case 'info': {

                response.data = { version: "0.11.0", activeServices: ["http", "kafka", "ldap", "mail", "mqtt"], search: { enabled: true } }
                return
            }
            case 'services':
                response.data = getServices(request.query['type'])
                return
            case 'serviceHttp':
                const name = request.path['name']
                const service = httpServices.find(x => x.name === name)
                if (service) {
                    response.data = {
                        name: service.name,
                        description: service.description,
                        version: service.version,
                        ...service.contact ? { contact: service.contact } : {},
                        servers: service.servers,
                        paths: service.paths.map(p => {
                            return {
                                path: p.path,
                                ...p.summary ? { summary: p.summary } : {},
                                ...p.description ? { description: p.description } : {},
                                status: p.status,
                                ...p.errors ? { errors: p.errors } : {},
                                operations: p.operations.map(o => {
                                    return {
                                        method: o.method,
                                        ...o.summary ? {summary: o.summary} : {},
                                        ...o.description ? {description: o.description} : {},
                                        ...o.operationId ? { operationId: o.operationId } : {},
                                        deprecated: o.deprecated ?? false,
                                        ...o.tags ? { tags: o.tags } : {},
                                        status: o.status,
                                        ...o.errors ? { errors: o.errors } : {},
                                        metrics: {
                                            http_requests_total: sum('http_requests_total', { name: 'service', value: service.name }, { name: 'endpoint', value: p.path}, { name: 'method', value: o.method.toUpperCase() }),
                                            http_requests_errors_total: sum('http_requests_errors_total', { name: 'service', value: service.name }, { name: 'endpoint', value: p.path}, { name: 'method', value: o.method.toUpperCase() }),
                                            http_request_timestamp: max('http_request_timestamp', { name: 'service', value: service.name }, { name: 'endpoint', value: p.path}, { name: 'method', value: o.method.toUpperCase() })
                                        }
                                    }
                                })
                            }
                        }),
                    }
                } else{
                    response.statusCode = 404
                }
                return
            case 'httpOperations': {
                const name = request.path['name']
                const service = httpServices.find(x => x.name === name)
                if (!service) {
                    response.statusCode = 404
                    return
                }
                const method = request.query['method']
                const path = request.query['path']
                let operations = []
                for (const p of service.paths) {
                    if (!path || p.path === path) {
                        for (const o of p.operations) {
                            if (!method || o.method.toLowerCase() === method.toLowerCase()) {
                                operations.push({
                                    ...o,
                                    metrics: {
                                        http_requests_total: sum('http_requests_total', { name: 'service', value: service.name }, { name: 'endpoint', value: p.path}, { name: 'method', value: o.method.toUpperCase() }),
                                        http_requests_errors_total: sum('http_requests_errors_total', { name: 'service', value: service.name }, { name: 'endpoint', value: p.path}, { name: 'method', value: o.method.toUpperCase() }),
                                        http_request_timestamp: max('http_request_timestamp', { name: 'service', value: service.name }, { name: 'endpoint', value: p.path}, { name: 'method', value: o.method.toUpperCase() })
                                    }
                                })
                            }
                        }
                    }
                }
                response.data = operations
                return
            }
            case 'serviceKafka':
                response.data = {
                    name: kafkaClusters[0].name,
                    description: kafkaClusters[0].description,
                    version: kafkaClusters[0].version,
                    contact: kafkaClusters[0].contact,
                    servers: kafkaClusters[0].servers,
                    topics: kafkaClusters[0].topics.map(x => {
                        return {
                            name: x.name,
                            ...x.summary === undefined ? {} : { summary: x.summary },
                            metrics: {
                                kafka_messages_total: metrics.find(m => m.name === `kafka_messages_total{service="Kafka World",topic="${x.name}"}`).value,
                                kafka_message_timestamp: metrics.find(m => m.name === `kafka_message_timestamp{service="Kafka World",topic="${x.name}"}`).value
                            }
                        }
                    }),
                    groups: kafkaClusters[0].groups.map(x => {
                        return {
                            name: x.name,
                            generation: x.generation,
                            state: x.state,
                            protocol: x.protocol,
                            members: x.members.length ?? 0,
                            metrics: x.metrics
                        }
                    }),
                    configs: kafkaClusters[0].configs.map(c => {
                        return {
                            id: c.id,
                            url: c.url,
                            provider: c.provider,
                            time: c.time
                        }
                    })
                }
                return
            case 'kafkaGroup':
                response.data = kafkaClusters[0].groups.find(x => x.name === request.path['group'])
                return
            case 'serviceMail':
                response.data = mailServices[0]
                return
            case 'kafkaClusters':
                response.data = kafkaClusters.map(x => getInfo(x, 'kafka'))
                return
            case 'kafkaTopics':
                response.data = kafkaClusters[0].topics
                return
            case 'kafkaTopic':
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                response.data = {
                    ...topic,
                    groups: kafkaClusters[0].groups.filter(g => g.topics.includes(topic.name)).map(x => {
                        return {
                            name: x.name,
                            generation: x.generation,
                            state: x.state,
                            protocol: x.protocol,
                            members: x.members.length ?? 0,
                            metrics: x.metrics
                        }
                    }),
                }
                if (response.data === undefined) {
                    response.statusCode = 404
                }
                return
            case 'kafkaProduceTopic': {
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                } else {
                    response.data = {
                        offsets: [
                            {
                                partition: 0,
                                offset: 1234,
                            }
                        ]
                    }
                }
                return
            }
            case 'kafkaPartitions': {
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                    return
                }
                response.data = topic.partitions
                return
            }
            case 'kafkaPartition': {
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                    return
                }
                const id = request.path.partitionId
                if (id >= topic.partitions.length) {
                    response.statusCode = 404
                } else {
                    response.data = topic.partitions[id]
                }
                return
            }
            case 'kafkaProducePartition': {
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                    return
                }
                const id = request.path.partitionId
                if (id >= topic.partitions.length) {
                    response.statusCode = 404
                } else {
                    response.data = {
                        offsets: [
                            {
                                partition: id,
                                offset: 1234,
                            }
                        ]
                    }
                }
                return
            }
            case 'kafkaOffsets': {
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                    return
                }
                const id = request.path.partitionId
                if (id >= topic.partitions.length) {
                    response.statusCode = 404
                } else {
                    const events = getEvents({topic: topic.name}).filter(x => x.data.partition === id)
                    if (request.header['Accept'] === 'application/json') {
                        response.data = events.map(x =>  ({
                            key: x.data.key.value,
                            value: JSON.parse(x.data.message.value)
                        }))
                    } else {
                        response.data = events.map(x => ({
                            key: base64.encode(x.data.key.value),
                            value: base64.encode(x.data.message.value)
                        }))
                    }
                }
                return
            }
            case 'kafkaOffset': {
                const topic = kafkaClusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                    return
                }
                const id = request.path.partitionId
                if (id >= topic.partitions.length) {
                    response.statusCode = 404
                } else {
                    const events = getEvents({topic: topic.name}).filter(x => x.data.partition === id && x.data.offset === request.path.offset)
                    if (events.length === 0) {
                        response.statusCode = 404
                        return
                    }
                    if (request.header['Accept'] === 'application/json') {
                        response.data = {
                            key: events[0].data.key.value,
                            value: JSON.parse(events[0].data.message.value)
                        }
                    } else {
                        response.data = {
                            key: base64.encode(events[0].data.key.value),
                            value: base64.encode(events[0].data.message.value)
                        }
                    }
                }
                return
            }
            case 'mailboxes':
                response.data = mailServices[0].mailboxes
                return
            case 'mailbox':
                for (const mb of mailServices[0].mailboxes) {
                    if (mb.name === request.path.mailbox) {
                        response.data = {
                            ...mb,
                            folders: mb.folders.map(x => x.name)
                        }
                        return
                    }
                }
                response.statusCode = 404
                return
            case 'smtpMessages':
                for (const mb of mailServices[0].mailboxes) {
                    if (mb.name === request.path.mailbox) {
                        const result = []
                        for (const f of mb.folders) {
                            result.push(...f.messages)
                        }
                        response.data = result
                        return
                    }
                }
                response.statusCode = 404
                return
            case 'smtpMail':
            case 'smtpMessage':
                const messageId = request.path['messageId']
                const mail = getMail(messageId)
                if (mail == null) {
                    response.statusCode = 404
                } else {
                    response.data = {
                        service: mailServices[0].name,
                        data: mail
                    }
                }
                return
            case 'smtpMailAttachment':
            case 'smtpAttachment':
                const attachment = getAttachment(request.path['messageId'], request.path['name'])
                if (attachment == null) {
                    response.statusCode = 404
                } else {
                    response.data = attachment.data
                    response.headers['Content-Type'] = attachment.contentType
                }
                return
            case 'serviceLdap':
                response.data = ldapServers[0]
                return
            case 'metrics':
                let names = request.query['names']
                if (!names) {
                    response.data = metrics
                } else{
                    response.data = metrics.filter(x => names.includes(x.name))
                }
                return
            case 'metricsType':
                let type = request.path['type']
                response.data = metrics.filter(x => x.name.startsWith(type + '_'))
                return
            case 'events':
                response.data = getEvents(request.query['traits'])
                return
            case 'event':
                let id = request.path["id"]
                let e = getEvent(id)
                if (!e){
                    response.statusCode = 404
                }else{
                    response.data = e
                }
                return
            case 'example': {
                const res = post(`${apiBaseUrl}/api/schema/example`, request.body)
                response.statusCode = res.statusCode
                response.headers = res.headers
                response.body = res.body
                return
            }
            case 'validate':
                sleep(1000)
                let url = `${apiBaseUrl}/api/schema/validate`
                if (request.query['outputFormat']) {
                    url += '?outputFormat=' + request.query['outputFormat']
                }
                const res = post(url, request.body)
                console.log(res.statusCode)
                response.statusCode = res.statusCode
                response.headers = res.headers
                response.body = res.body
                response.data = null
                return
            case 'configs':
                response.data = getConfigs()
                return
            case 'config':
                const config = getConfig(request.path.id)
                if (config) {
                    response.data = config
                    return
                } else {
                    response.statusCode = 404
                    response.data = ''
                    return
                }
            case 'configData':
                const configData = configs[request.path.id]
                if (configData) {
                    response.data = configData.data
                    response.headers = configData.headers
                    return
                } else {
                    response.statusCode = 404
                    response.data = ''
                    return
                }
            case 'fakerTree':
                const resp = get(`${apiBaseUrl}/mokapi/api/faker/tree`)
                response.body = resp.body
                return
            case 'search':
                switch (request.query.q) {
                    case 'bad':
                        response.statusCode = 400
                        response.data = { message: 'invalid query parameter \'limit\': must be a number'}
                        break
                    case 'empty':
                        response.data = { results: [], total: 0}
                        break
                    default:
                        response.data = getSearchResults()
                        break
                }
        }
    }, {tags: {name: "dashboard"}})

    for (const config of [...Object.values(httpConfigs),...Object.values(kafkaConfigs)]) {
        const r = await fetch(config.data)
        configs[config.id] = {
            data:   r.body,
            headers: {
                'Content-Disposition': `inline; filename="${config.filename}"`
            }
        }
    }

    jobs.push(
        {
            id: "8234",
            traits: {
                namespace: "job",
                name: "cron",
            },
            time: new Date(Date.now()).toISOString(),
            data: {
                schedule: '3s',
                maxRuns: -1,
                runs: 1,
                nextRun: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
                duration: 2300,
                tags: {
                    name: "cron",
                    file: "cron.js",
                    fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
                }
            }
        },
        {
            id: "8134",
            traits: {
                namespace: "job",
                name: "cron",
            },
            time: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
            data: {
                schedule: '3s',
                maxRuns: 10,
                runs: 1,
                nextRun: new Date(Date.now()).toISOString(),
                duration: 2300,
                logs: [
                    {
                        level: 'error',
                        message: 'an error message, see https://mokapi.io'
                    }
                ],
                error: {
                    message: 'There was an error'
                },
                tags: {
                    name: "cron",
                    file: "cron.js",
                    fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
                }
            },

        },
        {
            id: "81344",
            traits: {
                namespace: "job",
                name: "cron",
            },
            time: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
            data: {
                schedule: '3s',
                runs: 1,
                nextRun: new Date(Date.now()).toISOString(),
                duration: 2300,
                logs: [
                    {
                        level: 'warn',
                        message: 'a warning message, see https://mokapi.io'
                    }
                ],
                tags: {
                    name: "cron",
                    file: "cron.js",
                    fileKey: "6f5cab53-9f51-4bbe-bb32-fe41ee71d542",
                }
            },

        }
        )
}

function getServices(type) {
    const services = []
    if (!type || type === 'http') {
        services.push(...httpServices.map(x => {
            return {
                ...getInfo(x, 'http'),
                metrics: {
                    http_requests_total: sum('http_requests_total', { name: 'service', value: x.name }),
                    http_requests_errors_total: sum('http_requests_errors_total', { name: 'service', value: x.name }),
                    http_request_timestamp: max('http_request_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'kafka') {
        services.push(...kafkaClusters.map(x => {
            return {
                ...getInfo(x, 'kafka'),
                metrics: {
                    kafka_messages_total: sum('kafka_messages_total', { name: 'service', value: x.name }),
                    kafka_message_timestamp: max('kafka_message_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'mail') {
        services.push(...mailServices.map(x => {
            return {
                ...getInfo(x, 'mail'),
                metrics: {
                    mail_mails_total: sum('mail_mails_total', { name: 'service', value: x.name }),
                    mail_mail_timestamp: max('mail_mail_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'ldap') {
        services.push(...ldapServers.map(x => {
            return {
                ...getInfo(x, 'ldap'),
                metrics: {
                    ldap_requests_total: sum('ldap_requests_total', { name: 'service', value: x.name }),
                    ldap_request_timestamp: max('ldap_request_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }
    if (!type || type === 'mqtt') {
        services.push(...mqttClusters.map(x => {
            return {
                ...getInfo(x, 'mqtt'),
                metrics: {
                    mqtt_messages_total: sum('mqtt_messages_total', { name: 'service', value: x.name }),
                    mqtt_message_timestamp: max('mqtt_message_timestamp', { name: 'service', value: x.name })
                }
            }
        }))
    }

    return services
}

function getEvent(id) {
    for (const e of httpEvents){
        if (e.id === id){
            return e
        }
    }
    for (const e of kafkaEvents){
        if (e.id === id){
            return e
        }
    }
     for (const e of mqttEvents){
        if (e.id === id){
            return e
        }
    }
    for (const e of mailEvents){
        if (e.id === id){
            return e
        }
    }
    for (const e of searches){
        if (e.id === id){
            return e
        }
    }
    for (const e of jobs) {
        if (e.id === id) {
            return e
        }
    }
    return null
}

function getEvents(traits) {
    const result = []
    for (const e of httpEvents){
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (const e of kafkaEvents){
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (const e of mqttEvents){
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (const e of mailEvents) {
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (const e of searches) {
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (const e of jobs) {
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    return result
}

function matchEvent(e, traits) {
    for (let t in traits) {
        if (e.traits[t] !== traits[t]) {
            return false
        }
    }
    return true
}

function getConfig(id) {
    if (httpConfigs[id]) {
        return httpConfigs[id]
    }
    if (kafkaConfigs[id]) {
        return kafkaConfigs[id]
    }
    if (jobConfig[id]) {
        return jobConfig[id]
    }
    return null
}

function getConfigs() {
    return [
        ...Object.values(httpConfigs),
        ...Object.values(kafkaConfigs),
        ...Object.values(jobConfig)
    ]
}

function getInfo(config, type) {
    const info = {
        name: config.name,
        description: config.description,
        version: config.version,
        type: type,
    }
    if (config.contact) {
        info.contact = {
            name: config.contact.name,
            email: config.contact.email,
            url: config.contact.url
        }
    }
    if (config.status) {
        info.status = config.status
    }
    return info
}

const jobConfig = {
    '6f5cab53-9f51-4bbe-bb32-fe41ee71d542': {
        id: '6f5cab53-9f51-4bbe-bb32-fe41ee71d542',
        url: 'file://cron.js',
        provider: 'file',
        time: '2025-05-01T08:49:25.482366+01:00',
        data: "import { every } from 'mokapi'; export default function() { every('1s', function() {}); }",
        filename: 'cron.js'
    }
}

function getSearchResults() {
    return {
        results: [
            {
                type: 'HTTP',
                title: "Swagger Petstore",
                fragments: ['<mark>Everything</mark>', 'store'],
                params: {
                    type: 'http',
                    service: 'Swagger Petstore',
                 }
            },
            {
                type: 'HTTP',
                domain: 'Swagger Petstore',
                title: "/pet",
                fragments: ['<mark>Everything</mark>', 'store'],
                params: {
                    type: 'http',
                    service: 'Swagger Petstore',
                    path: '/pet',
                    methods: 'get,post'
                }
            },
            {
                type: 'HTTP',
                domain: 'Swagger Petstore',
                title: "/pet",
                fragments: ['<mark>Everything</mark>', 'store'],
                params: {
                    type: 'http',
                    service: 'Swagger Petstore',
                    path: '/pet',
                    method: 'post',
                    statusCode: '200'
                }
            },
             {
                type: 'HTTP',
                domain: 'Swagger Petstore',
                title: "/foo/longTextPathEndpoint/bar/pets",
                fragments: ['<mark>Everything</mark>', 'store'],
                params: {
                    type: 'http',
                    service: 'Swagger Petstore',
                    path: '/foo/bar/pets',
                    method: 'put',
                    statusCode: '200'
                }
            },
            {
                type: 'HTTP',
                domain: 'Swagger Petstore',
                title: "/customer/foo/bar/test/correspondenceAddress",
                fragments: ['<mark>Everything</mark>', 'store'],
                params: {
                    type: 'http',
                    service: 'Swagger Petstore',
                    path: '/customer/foo/bar/test/correspondenceAddress',
                    method: 'post',
                    statusCode: '200'
                }
            },
            {
                type: 'Event',
                domain: 'Swagger Petstore',
                title: "http://127.0.0.1:18080/pet",
                fragments: ['<mark>POST</mark>'],
                time: '2025-05-23T08:49:25.482366+01:00',
                params: {
                    type: 'event',
                    'traits.namespace': 'http',
                    'traits.name': 'Swagger Petstore',
                    'traits.path': '/pet',
                    'traits.method': 'POST',
                    id: '9e058601-562a-4063-a3b2-066ba624893a'
                }
            },
            {
                type: 'Config',
                domain: 'FILE',
                title: "file://root/path/to/my/foo/file.json",
                fragments: ['<mark>foo</mark> bar'],
                params: {
                    type: 'config',
                    id: 'b6fea8ac-56c7-4e73-a9c0-6887640bdca8'
                }
            },
            {
                type: 'KAFKA',
                title: "Kafka World",
                fragments: ['To ours <mark>significant</mark> why upon tomorrow'],
                params: {
                    type: 'kafka',
                    service: 'Kafka World',
                    topics: '3'
                }
            },
            {
                type: 'KAFKA',
                domain: 'Kafka World',
                title: "mokapi.shop.products",
                fragments: ['Though literature second <mark>anywhere fortnightly</mark>'],
                params: {
                    type: 'kafka',
                    service: 'Kafka World',
                    topic: 'mokapi.shop.products'
                }
            },
            {
                type: 'Event',
                domain: 'Kafka World',
                title: "GGOEWXXX0827",
                fragments: null,
                time: '2025-05-01T08:49:25.482366+01:00',
                params: {
                    type: 'event',
                    'traits.namespace': 'kafka',
                    'traits.name': 'Kafka World',
                    'traits.topic': 'mokapi.shop.products',
                    'traits.type': 'message',
                    'traits.partition': '0',
                    id: '5d549ffe-97cd-477d-8db6-27ed4963c08d'
                }
            },
            {
                type: 'Mail',
                domain: 'Mail Testserver',
                title: "Mail Testserver",
                fragments: ['This is a sample <mark>smtp</mark> server'],
                params: {
                    type: 'mail',
                    service: 'Mail Testserver',
                }
            },
            {
                type: 'Event',
                domain: 'Mail Testserver',
                title: "A test mail",
                fragments: ['<mark>test</mark> <mark>mail</mark>'],
                time: '2025-05-01T08:49:25.482366+01:00',
                params: {
                    type: 'event',
                    'traits.namespace': 'mail',
                    'traits.name': 'Mail Testserver',
                    id: '273cd167-f5a5-4da1-969e-d44213686491'
                }
            },
            {
                type: 'LDAP',
                title: "LDAP Testserver",
                fragments: ['This is a sample <mark>LDAP</mark> server'],
                params: {
                    type: 'ldap',
                    service: 'LDAP Testserver',
                    entries: '12'
                }
            },
            {
                type: 'Event',
                domain: 'LDAP Testserver',
                title: "(objectClass=user)",
                fragments: ['(objectClass=<mark>user</mark>)'],
                time: '2025-05-01T08:49:25.482366+01:00',
                params: {
                    type: 'event',
                    'traits.namespace': 'ldap',
                    'traits.name': 'LDAP Testserver',
                    'traits.operation': 'search',
                    id: 'dkads-23124',
                    baseDn: 'dc=example,dc=org',
                }
            }
        ],
        facets: {
            type: [{value: 'HTTP', count: 3}, {value: 'Kafka', count: 2}, {value: 'Mail', count: 1}, {value: 'Event', count: 2}, {value: 'Config', count: 1}]
        },
        total: 15
    }
}