import { app } from 'mokapi'

export default function() {
    app.http().post('/sim/websocket', (req, res) => {
        try {
            app.websocket().send('/dm', req.body.message, { messageId: req.body.messageId })
            res.statusCode = 204
        } catch (e) {
            res.statusCode = 422
            res.body = String(e)
        }
    })

    app.http().post('/sim/websocket-async', async (req, res) => {
        try {
            await app.websocket().sendAsync('/dm', req.body.message, { messageId: req.body.messageId })
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
