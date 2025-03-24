import { send } from 'mokapi/smtp'


export default function() {
    send(
        'smtp://127.0.0.1:8025',
        {
            from: {name: 'Alice', address: 'alice@mokapi.io'},
            to: ['bob@mokapi.io'],
            subject: 'A test mail',
            contentType: 'text/html',
            body: '<h1>Hello Bob</h1> How you\'re doing?'
        }
    )
}