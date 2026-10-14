import { on, sleep } from 'mokapi'
import { clusters as kafkaClusters, events as kafkaEvents, configs as kafkaConfigs } from './kafka.ts'
import { apps as httpServices, events as httpEvents, configs as httpConfigs } from './http'
import { services as mailServices, mailEvents, getMail, getAttachment } from './mail'
import { server as ldapServers, searches } from './ldap'
import { clusters as mqttClusters, events as mqttEvents } from './mqtt'
import { services as websocketServices, events as websocketEvents } from './websocket'
import { metrics, sum, max } from './metrics.ts'
import { get, post, fetch } from 'mokapi/http'

const configs = {}

export default async function() {
    const apiBaseUrl = `http://localhost:8091/mokapi`

    on('http', function(request, response) {

        switch (request.operationId) {
            case 'serviceMail':
                response.data = mailServices[0]
                return
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
            case 'fakerTree': {
                const resp = get(`${apiBaseUrl}/mokapi/api/faker/tree`)
                response.body = resp.body
                return
            }
            case 'fakerNodeFake': {
                const resp = get(`${apiBaseUrl}/mokapi/api/faker/node/${request.path.name}/fake`)
                response.statusCode = resp.statusCode
                response.body = resp.body
                return
            }
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
                description: 'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam...',
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
                    messageId: '20230223-084925.763-4196@mokapi.io'
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