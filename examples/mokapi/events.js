import {on} from 'mokapi'

export default function() {
    on('http', function(request, response) {
        response.headers["Access-Control-Allow-Origin"] = '*'
        if (request.operationId === 'events') {
            if (request.path["namespace"] === 'kafka') {
                response.data = kafka_events
            }
        }
    })
}

const kafka_events = [
    {
        "id": "123456",
        "traits": {
            "namespace": "kafka",
        },
        "data": {
            "offset": 0,
            "key": "foo",
            "message": "{\"id\": 12345}",
            "time": 1651771269
        }
    }
]