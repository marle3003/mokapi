import { on } from 'mokapi'

export default function() {
    on('websocket', (event) => {
        event.broadcast({ 
            from: event.message.username,
            text: event.message.text
        })
    }, { track: true })
}