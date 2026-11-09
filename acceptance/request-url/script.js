import { app } from 'mokapi'

export default function() {
    app.http().get('/url', (req, res) => {
        res.data = { scheme: req.url.scheme, url: req.url.toString() }
    })

    // registered last: it answers 'ready' once the script has run completely
    app.http().get('/sim/ready', (req, res) => {
        res.body = 'ready'
    })
}
