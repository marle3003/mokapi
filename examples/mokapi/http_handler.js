import {on} from 'mokapi'
import {clusters, events as kafkaEvents} from 'kafka.js'
import {apps as httpServices, events as httpEvents} from 'services_http.js'
import {server as smtpServers, mails} from 'smtp.js'
import {metrics} from 'metrics.js'
import {fake} from 'mokapi/faker'

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Methods"] = "*"
        response.headers["Access-Control-Allow-Headers"] = "*"
        response.headers["Access-Control-Allow-Origin"] = "*"

        switch (request.operationId) {
            case 'info':
                response.data = {version: "0.5.0", activeServices: ["http", "kafka", "smtp"]}
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
                console.log(request.body.properties.id)
                response.data = fake(request.body)
                return true
        }
    }, {tags: {name: "dashboard"}})
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

    return http.concat(kafka).concat(smtp)
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
    for (let e of mails) {
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
