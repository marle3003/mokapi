import { on, app } from 'mokapi'
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

    app.http()
        .route('/pet/findByStatus')
        .use((req, res) => {
            res.ctx.name = 'Zoe'
        }, { tags: { middleware: '' }})
        .get((req, res) => {
            res.data = [
                {
                    name: res.ctx.name,
                    photoUrls: []
                }
            ]
        })
    app.http().get('/pet/findByTags', (req, res) => {
        res.data = [
            {
                name: 'Chili',
                photoUrls: []
            }
        ]
    }, {
        // this handler should not be called on findByStatus
        // value -1 ensures the handler is called at last
        // However, if the handler is executed anyway, the response would be overwritten.
        priority: -1
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