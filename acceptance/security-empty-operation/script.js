import { app } from 'mokapi'

export default function() {
    // answers with the Authorization header the mock exposes to the script
    for (const path of ['/protected', '/public']) {
        app.http().get(path, (req, res) => {
            res.data = req.header['Authorization'] ?? 'none'
        })
    }

    // registered last: it answers 'ready' once the script has run completely
    app.http().get('/sim/ready', (req, res) => {
        res.body = 'ready'
    })
}
