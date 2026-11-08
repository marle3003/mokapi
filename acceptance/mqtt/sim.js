import { app } from 'mokapi'

export default function() {
    app.http().post('/sim/mqtt', (req, res) => {
        try {
            app.mqtt().publish(req.body.topic, Object.assign({ retain: true }, req.body.message))
            res.statusCode = 204
        } catch (e) {
            res.statusCode = 422
            res.body = String(e)
        }
    })

    // registered last: it answers 'ready' once the script has run completely
    app.http().get('/sim/ready', (req, res) => {
        res.body = 'ready'
    })
}
