import { on } from 'mokapi'
import { clusters, events as kafkaEvents, configs as kafkaConfigs } from 'kafka.js'
import { apps as httpServices, events as httpEvents, configs as httpConfigs } from 'services_http.js'
import { server as smtpServers, mails, mailEvents, getMail, getAttachment } from 'smtp.js'
import { server as ldapServers, searches } from 'ldap.js'
import { metrics } from 'metrics.js'
import { fake } from 'mokapi/faker'
import { get } from 'mokapi/http'

const configs = {}

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Methods"] = "*"
        response.headers["Access-Control-Allow-Headers"] = "*"
        response.headers["Access-Control-Allow-Origin"] = "*"

        switch (request.operationId) {
            case 'info':
                response.data = {version: "0.5.0", activeServices: ["http", "kafka", "ldap", "smtp"]}
                return true
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
            case 'example':
                response.data = fake(request.body)
                return true
            case 'config':
                const config = getConfig(request.path.id)
                if (config) {
                    response.data = config
                    return true
                } else {
                    console.log("config not found: "+request.path.id)
                    response.statusCode = 404
                    response.data = ''
                    return true
                }
            case 'configdata':
                const configData = configs[request.path.id]
                if (configData) {
                    response.data = configData
                    return true
                } else {
                    console.log("config not found: "+JSON.stringify(configs))
                    response.statusCode = 404
                    response.data = ''
                    return true
                }
        }
    }, {tags: {name: "dashboard"}})

    for (const config of [...Object.values(httpConfigs),...Object.values(kafkaConfigs)]) {
        const r = get(config.data)
        configs[config.id] = r.body
    }
}

function getServices() {
    let http = httpServices.map(x => {
        x.type = "http"
        return x
    })
    let kafka = clusters.map(x => {
        let c = {...x}
        c.type = "kafka"
        c.topics = c.topics.map(t => t.name)
        return c
    })
    let smtp = smtpServers.map(x => {
        x.type = "smtp"
        return x
    })
    let ldap = ldapServers.map(x => {
        x.type = "ldap"
        return x
    })

    return http.concat(kafka).concat(smtp).concat(ldap)
}

function getEvent(id) {
    for (let e of httpEvents){
        if (e.id === id){
            return e
        }
    }
    for (let e of kafkaEvents){
        if (e.id === id){
            return e
        }
    }
    for (let e of mailEvents){
        if (e.id === id){
            return e
        }
    }
    for (let e of searches){
        if (e.id === id){
            return e
        }
    }
    return null
}

function getEvents(traits) {
    let result = []
    for (let e of httpEvents){
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (let e of kafkaEvents){
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (let e of mailEvents) {
        if (matchEvent(e, traits)) {
            result.push(e)
        }
    }
    for (let e of searches) {
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
    return null
}
