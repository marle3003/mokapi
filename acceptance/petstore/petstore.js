import {on} from 'mokapi'
import kafka from 'mokapi/kafka'

export default function() {
    on('kafka', function (record) {
        record.headers = []
        record.headers.push({key: 'foo', value: 'bar'})
        console.log(record)
    })
    kafka.produce({topic: 'petstore.order-event'})
    on('http', function(request, response) {
        if (request.operationId === "getPetById") {
            switch (request.path.petId) {
                case 2:
                    response.statusCode = 404
                    return true
                case 3:
                    response.statusCode = 404
                    response.data = null
                    return true
                case 4:
                    response.statusCode = 200
                    response.data = {

                    }
            }
        }
        return false
    });
}