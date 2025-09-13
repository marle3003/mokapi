import { on, sleep } from 'mokapi'
import { clusters, events as kafkaEvents, configs as kafkaConfigs } from 'kafka.js'
import { apps as httpServices, events as httpEvents, configs as httpConfigs } from 'services_http.js'
import { services as mailServices, mailEvents, getMail, getAttachment } from 'mail.js'
import { server as ldapServers, searches } from 'ldap.js'
import { metrics } from 'metrics.js'
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

                response.data = { version: "0.11.0", activeServices: ["http", "kafka", "ldap", "mail"], search: { enabled: true } }
                return true
            }
            case 'services':
                response.data = getServices()
                return true
            case 'serviceHttp':
                response.data = httpServices[0]
                return true
            case 'serviceKafka':
                response.data = clusters[0]
                return true
            case 'serviceMail':
                response.data = mailServices[0]
                return true
            case 'clusters':
                response.data = clusters.map(x => getInfo(x, 'kafka'))
                return
            case 'topics':
                response.data = clusters[0].topics
                return
            case 'topic':
                response.data = clusters[0].topics.find(x => x.name === request.path.topic)
                if (response.data === undefined) {
                    response.statusCode = 404
                }
                return
            case 'produceTopic': {
                const topic = clusters[0].topics.find(x => x.name === request.path.topic)
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
            case 'partitions': {
                const topic = clusters[0].topics.find(x => x.name === request.path.topic)
                if (!topic) {
                    response.statusCode = 404
                    return
                }
                response.data = topic.partitions
                return
            }
            case 'partition': {
                const topic = clusters[0].topics.find(x => x.name === request.path.topic)
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
            case 'producePartition': {
                const topic = clusters[0].topics.find(x => x.name === request.path.topic)
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
            case 'offsets': {
                const topic = clusters[0].topics.find(x => x.name === request.path.topic)
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
            case 'offset': {
                const topic = clusters[0].topics.find(x => x.name === request.path.topic)
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
                return true
            case 'mailbox':
                for (const mb of mailServices[0].mailboxes) {
                    if (mb.name === request.path.mailbox) {
                        response.data = {
                            ...mb,
                            folders: mb.folders.map(x => x.name)
                        }
                        return true
                    }
                }
                response.statusCode = 404
                return true
            case 'smtpMessages':
                for (const mb of mailServices[0].mailboxes) {
                    if (mb.name === request.path.mailbox) {
                        const result = []
                        for (const f of mb.folders) {
                            result.push(...f.messages)
                        }
                        response.data = result
                        return true
                    }
                }
                response.statusCode = 404
                return true
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
                return true
            case 'smtpMailAttachment':
            case 'smtpAttachment':
                const attachment = getAttachment(request.path['messageId'], request.path['name'])
                if (attachment == null) {
                    response.statusCode = 404
                } else {
                    response.data = attachment.data
                    response.headers['Content-Type'] = attachment.contentType
                }
                return true
            case 'serviceLdap':
                response.data = ldapServers[0]
                return true
            case 'metrics':
                let q = request.query["query"]
                if (!q) {
                    response.data = metrics
                } else{
                    response.data = metrics.filter(x => x.name.includes(q))
                }
                return true
            case 'events':
                response.data = getEvents(request.query['traits'])
                return true
            case 'event':
                let id = request.path["id"]
                let e = getEvent(id)
                if (!e){
                    response.statusCode = 404
                }else{
                    response.data = e
                }
                return true
            case 'example': {
                const res = post(`${apiBaseUrl}/api/schema/example`, request.body)
                response.statusCode = res.statusCode
                response.headers = res.headers
                response.body = res.body
                return true
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
                return true
            case 'configs':
                response.data = getConfigs()
                return true
            case 'config':
                const config = getConfig(request.path.id)
                if (config) {
                    response.data = config
                    return true
                } else {
                    response.statusCode = 404
                    response.data = ''
                    return true
                }
            case 'configData':
                const configData = configs[request.path.id]
                if (configData) {
                    response.data = configData.data
                    response.headers = configData.headers
                    return true
                } else {
                    response.statusCode = 404
                    response.data = ''
                    return true
                }
            case 'fakerTree':
                const resp = get(`${apiBaseUrl}/mokapi/api/faker/tree`)
                response.body = resp.body
                return true
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

function getServices() {
    let http = httpServices.map(x => getInfo(x, 'http'))
    let kafka = clusters.map(x => getInfo(x, 'kafka'))
    let smtp = mailServices.map(x => getInfo(x, 'mail'))
    let ldap = ldapServers.map(x => getInfo(x, 'ldap'))

    return http.concat(kafka).concat(smtp).concat(ldap)
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
        metrics: config.metrics,
        type: type
    }
    if (config.contact) {
        info.contact = {
            name: config.contact.name,
            email: config.contact.email,
            url: config.contact.url
        }
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
                }
            },
            {
                type: 'HTTP',
                domain: 'Swagger Petstore',
                title: "POST  /pet",
                fragments: ['<mark>Everything</mark>', 'store'],
                params: {
                    type: 'http',
                    service: 'Swagger Petstore',
                    path: '/pet',
                    method: 'post'
                }
            },
            {
                type: 'Event',
                domain: 'Swagger Petstore',
                title: "POST http://127.0.0.1:18080/pet",
                fragments: [],
                params: {
                    type: 'event',
                    namespace: 'http',
                    id: '4242'
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
                    service: 'Kafka World'
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
                domain: 'Kafka World - mokapi.shop.products',
                title: "GGOEWXXX0827",
                fragments: null,
                params: {
                    type: 'event',
                    namespace: 'kafka',
                    id: '123456'
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
                fragments: ['message <mark>from Alice</mark>'],
                params: {
                    type: 'event',
                    namespace: 'mail',
                    id: '8832'
                }
            }
        ],
        facets: {
            type: [{value: 'HTTP', count: 3}, {value: 'Kafka', count: 2}, {value: 'Mail', count: 1}, {value: 'Event', count: 2}, {value: 'Config', count: 1}]
        },
        total: 8
    }
}