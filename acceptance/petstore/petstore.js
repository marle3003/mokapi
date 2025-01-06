import {on} from 'mokapi'
import kafka from 'mokapi/kafka'

export default async function() {
    on('kafka', function (record) {
        record.headers = { foo: 'bar' }
    })
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
                    return true
                case 5:
                    // use generated data but change pet's name
                    response.data.name = 'Zoe'
                    return true
            }
        }
        return false
    });
    await kafka.produceAsync({ topic: 'petstore.order-event', messages: [{partition: 0}] })
}