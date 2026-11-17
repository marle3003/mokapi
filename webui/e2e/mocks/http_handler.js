import { on } from 'mokapi'
import { configs as kafkaConfigs } from './kafka.ts'
import { configs as httpConfigs } from './http'
import { fetch } from 'mokapi/http'

const configs = {}

export default async function() {
    on('http', function(request, response) {

        switch (request.operationId) {
            case 'search':
                switch (request.query.q) {
                    case 'bad':
                        response.statusCode = 400
                        response.data = { message: 'invalid query parameter \'limit\': must be a number'}
                        break
                    case 'empty':
                        response.data = { results: [], total: 0}
                        break
                    case 'many':
                        response.data = getManySearchResults(request.query)
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

const manyTotal = 57
const manyTypes = ['HTTP', 'Kafka', 'Mail', 'LDAP', 'Config', 'Event']
const manyEventKinds = ['http', 'kafka', 'mail', 'ldap', 'httpLog']
const manyStatusCodes = ['200', '201', '404', '500']

// q=many: 57 deterministic results, cycling through all six types; pages by `index` and `limit`
// (from = index * limit) and is restricted by the facet `type`, like api/handler_search.go.
// The facet counts ignore the type filter, so the select keeps all its options.
function getManySearchResults(query) {
    const all = Array.from({ length: manyTotal }, (_, i) => manyResult(i))
    const limit = query.limit ? parseInt(query.limit) : 10
    const index = query.index ? parseInt(query.index) : 0
    const matching = query.type ? all.filter(r => r.type === query.type) : all
    return {
        results: matching.slice(index * limit, (index + 1) * limit),
        facets: { type: manyTypes.map(value => ({ value, count: all.filter(r => r.type === value).length })) },
        total: matching.length
    }
}

function manyResult(i) {
    const type = manyTypes[i % manyTypes.length]
    const n = Math.floor(i / manyTypes.length) + 1
    const time = new Date(Date.UTC(2025, 4, 23, 8, 0, i)).toISOString()
    const fragments = [`match <mark>many</mark> ${i + 1}`]
    switch (type) {
        case 'HTTP':
            return {
                type, domain: 'Swagger Petstore', title: `/pet/${i + 1}`, fragments,
                params: {
                    type: 'http', service: 'Swagger Petstore', path: `/pet/${i + 1}`,
                    method: i % 2 === 0 ? 'get' : 'post', statusCode: manyStatusCodes[i % manyStatusCodes.length]
                }
            }
        case 'Kafka':
            return {
                type, domain: 'Kafka World', title: `mokapi.shop.topic${n}`, fragments,
                params: { type: 'kafka', service: 'Kafka World', topic: `mokapi.shop.topic${n}` }
            }
        case 'Mail':
            return {
                type, domain: 'Mail Testserver', title: `Mailbox ${n}`, fragments,
                params: { type: 'mail', service: 'Mail Testserver', mailbox: `mailbox${n}` }
            }
        case 'LDAP':
            return {
                type, domain: 'LDAP Testserver', title: `LDAP Testserver ${n}`, fragments,
                params: { type: 'ldap', service: 'LDAP Testserver' }
            }
        case 'Config':
            return {
                type, domain: 'FILE', title: `file://root/path/to/config${n}.json`, fragments,
                params: { type: 'config', id: `config-${n}` }
            }
        default:
            return manyEvent(i, n, time, fragments)
    }
}

function manyEvent(i, n, time, fragments) {
    const kind = manyEventKinds[n % manyEventKinds.length]
    const id = `event-${i + 1}`
    switch (kind) {
        case 'http':
            return {
                type: 'Event', domain: 'Swagger Petstore', title: `http://127.0.0.1:18080/pet/${i + 1}`, fragments, time,
                params: {
                    type: 'event', 'traits.namespace': 'http', 'traits.name': 'Swagger Petstore', 'traits.method': 'POST',
                    'traits.path': `/pet/${i + 1}`, statusCode: '500', id
                }
            }
        case 'kafka':
            return {
                type: 'Event', domain: 'Kafka World', title: `KEY${i + 1}`, fragments, time,
                params: {
                    type: 'event', 'traits.namespace': 'kafka', 'traits.name': 'Kafka World',
                    'traits.topic': 'mokapi.shop.products', 'traits.type': 'message', 'traits.partition': '0', id
                }
            }
        case 'mail':
            return {
                type: 'Event', domain: 'Mail Testserver', title: `Mail ${i + 1}`, fragments, time,
                params: { type: 'event', 'traits.namespace': 'mail', 'traits.name': 'Mail Testserver', messageId: `${id}@mokapi.io` }
            }
        case 'ldap':
            return {
                type: 'Event', domain: 'LDAP Testserver', title: `(objectClass=user${i + 1})`, fragments, time,
                params: {
                    type: 'event', 'traits.namespace': 'ldap', 'traits.name': 'LDAP Testserver',
                    'traits.operation': 'search', baseDn: 'dc=example,dc=org', id
                }
            }
        default:
            return {
                type: 'Event', domain: 'Swagger Petstore', title: `script error ${i + 1}`, fragments, time,
                params: {
                    type: 'event', 'traits.namespace': 'http', 'traits.name': 'Swagger Petstore', 'traits.type': 'log', id
                }
            }
    }
}
