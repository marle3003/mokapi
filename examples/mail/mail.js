import { send } from 'mokapi/smtp'


export default function() {
    try {
        send(
            'smtp://localhost:25',
            {
                from: {name: 'Alice', address: 'alice@mokapi.io'},
                to: ['bob@mokapi.io'],
                subject: 'A test mail',
                contentType: 'text/html',
                body: '<h1>Hello Bob</h1> How you\'re doing?'
            },
            {
                plain: {
                    username: 'alice@mokapi.io',
                    password: ''
                }
            }
        )
    }catch (e) {
        console.error(e)
    }
}