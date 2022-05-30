import {on} from 'mokapi'
import {clusters, events as kafkaEvents} from 'kafka.js'
import {apps as httpServices, events as httpEvents} from 'services_http'
import {metrics} from 'metrics'

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Origin"] = '*'

        switch (request.operationId) {
            case 'info':
                response.data = {version: "1.0", activeServices: ["http", "kafka"]}
                return true
            case 'services':
                response.data = getServices()
                return true
            case 'serviceKafka':
                console.log(clusters)
                response.data = clusters[0]
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

    return http.concat(kafka)
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
