import { on, env, sleep } from 'mokapi'
import { clusters, events as kafkaEvents, configs as kafkaConfigs } from 'kafka.js'
import { apps as httpServices, events as httpEvents, configs as httpConfigs } from 'services_http.js'
import { server as smtpServers, mails, mailEvents, getMail, getAttachment } from 'smtp.js'
import { server as ldapServers, searches } from 'ldap.js'
import { metrics } from 'metrics.js'
import { get, post, fetch } from 'mokapi/http'

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

                response.data = {version: "0.11.0", activeServices: ["http", "kafka", "ldap", "smtp"]}
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
            case 'serviceSmtp':
                response.data = smtpServers[0]
                return true
            case 'smtpMail':
                const messageId = request.path['messageId']
                const mail = getMail(messageId)
                if (mail == null) {
                    response.statusCode = 404
                } else {
                    response.data = mail
                }
                return true
            case 'smtpMailAttachment':
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
            case 'configdata':
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
            case 'fakertree':
                const resp = get(`${apiBaseUrl}/mokapi/api/faker/tree`)
                response.body = resp.body
                return true
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
    let smtp = smtpServers.map(x => getInfo(x, 'smtp'))
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
        config.contact = {
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
