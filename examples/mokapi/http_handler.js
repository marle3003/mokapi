import {on} from 'mokapi'
import {clusters} from 'kafka'
import {apps as httpServices} from 'services_http'

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Origin"] = '*'

        switch (request.operationId) {
            case "services":
                response.data = httpServices
        }

        if (request.operationId === 'serviceKafka') {
            response.data = clusters[0]
        }
    })
}