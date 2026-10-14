import { on } from 'mokapi'

export default function() {
    on('websocket', (event) => {
        if (event.type === 'message') {
            event.broadcast({ 
                from: event.message.username,
                text: event.message.text
            })
        }
    }, { track: true })
}