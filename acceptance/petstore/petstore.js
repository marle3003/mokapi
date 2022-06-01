import {on} from 'mokapi'
import kafka from 'kafka'

export default function() {
    kafka.produce({topic: 'petstore.order-event'})
    on('http', function(request, response) {
        if (request.operationId === "getPetById") {
            if (request.path.petId === 2) {
                response.statusCode = 404
                return true
            }else if (request.path.petId === 3) {
                response.statusCode = 404
                response.data = null
                return true
            }
        }
        return false
    });
}