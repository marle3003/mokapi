import { on } from 'mokapi'

export default function() {
    on('websocket', (event) => {
        event.reply({ text: 'pong' })
    })
}