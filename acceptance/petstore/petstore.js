import {on} from 'mokapi'
import kafka from 'mokapi/kafka'

export default async function() {
    on('kafka', function (record) {
        record.headers = { foo: 'bar', schemaId: record.schemaId }
    })
    on('http', function(request, response) {
        if (request.operationId === "getPetById") {
            switch (request.path.petId) {
                case 2:
                    response.statusCode = 404
                    break
                case 3:
                    response.statusCode = 404
                    response.data = null
                    break
                case 4:
                    response.statusCode = 200
                    response.data = {

                    }
                    break
                case 5:
                    // use generated data but change pet's name
                    response.data.name = 'Zoe'
                    break
            }
        }
    });
    await kafka.produceAsync({
        topic: 'petstore.order-event',
        cluster: 'A sample AsyncApi Kafka streaming api',
        messages: [{partition: 0}]
    })
    await kafka.produceAsync({
        topic: 'petstore.order-event',
        cluster: 'Petstore Stream API',
    })
}