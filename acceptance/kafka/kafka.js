import { on } from 'mokapi'
import kafka from 'mokapi/kafka'

export default async function() {
    on('kafka', function (record) {
        record.headers = { foo: 'bar', schemaId: record.schemaId }
    })

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