import { app, on } from 'mokapi'

// every mail a handler has seen, read by the acceptance test
const received = []

function addresses(list) {
    return (list || []).map(a => a.address)
}

function record(handler, mail) {
    received.push({
        handler,
        api: mail.api,
        subject: mail.subject,
        messageId: mail.messageId,
        time: mail.time instanceof Date ? mail.time.toISOString() : null,
        encoding: mail.encoding,
        to: addresses(mail.to),
        cc: addresses(mail.cc),
        bcc: addresses(mail.bcc),
    })
}

const isTracked = mail => mail.subject.startsWith('track:')

export default function() {
    app.http().get('/received', (req, res) => {
        res.data = received
    })

    app.http().post('/send', async (req, res) => {
        const { scope, async: useAsync, message } = req.body
        try {
            const router = scope ? app.api(scope).mail() : app.mail()
            if (useAsync) {
                const result = await router.sendAsync(message)
                res.data = { ok: true, resolvesRouter: typeof result.send === 'function' }
            } else {
                router.send(message)
                res.data = { ok: true }
            }
        } catch (e) {
            res.data = { ok: false, error: String(e) }
        }
    })

    app.mail().message(mail => record('all', mail))
    app.api('Mail B').mail().message(mail => record('api-b', mail))
    on('smtp', mail => record('on-smtp', mail))

    app.mail().mailbox('alice@a.io').message(mail => record('alice', mail))
    app.mail().mailbox('Alice@A.io').message(mail => record('alice-upper', mail))
    app.mail().mailbox('bob@a.io').message(mail => record('bob', mail))
    app.mail().mailbox('carol@a.io').message(mail => record('carol', mail))
    app.api('Mail B').mail().mailbox('alice@b.io').message(mail => record('b-alice', mail))

    // runs first: the changed subject is what gets stored, the envelope stays
    app.mail().message(mail => {
        if (mail.subject.startsWith('rewrite:')) {
            mail.subject = 'rewritten ' + mail.subject
            mail.to = [{ address: 'other@a.io' }]
        }
    }, { priority: 100 })

    // registered in the opposite order of their priority
    app.mail().message(() => {}, { tags: { order: 'second' }, track: isTracked })
    app.mail().message(() => {}, { tags: { order: 'first' }, priority: 10, track: isTracked })

    // registered last: it answers 'ready' once the script has run completely
    app.http().get('/sim/ready', (req, res) => {
        res.body = 'ready'
    })
}
