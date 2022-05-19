import {on} from 'mokapi'
import {clusters} from 'kafka.js'
import {apps as httpServices, events as httpEvents} from 'services_http'
import {metrics} from 'metrics'

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Origin"] = '*'

        switch (request.operationId) {
            case 'info':
                response.data = {version: "1.0"}
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
                switch (request.path["namespace"]) {
                    case "http":
                        response.data = httpEvents
                        return true
                }
                return false
        }
    })
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