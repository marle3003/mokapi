import {on} from 'mokapi'
import {clusters} from 'kafka'
import {apps as httpServices} from 'services_http'

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Origin"] = '*'

        switch (request.operationId) {
            case "services":
                response.data = getServices()
        }

        if (request.operationId === 'serviceKafka') {
            response.data = clusters[0]
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