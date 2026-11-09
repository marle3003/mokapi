import { app, delay } from 'mokapi'

export default function() {
    app.http().get('/slow', async (req, res) => {
        await delay('2s')
        res.body = 'slow'
    })

    app.http().get('/fast', (req, res) => {
        res.body = 'fast'
    })

    // registered last: it answers 'ready' once the script has run completely
    app.http().get('/sim/ready', (req, res) => {
        res.body = 'ready'
    })
}
