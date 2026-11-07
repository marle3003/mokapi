import { app } from 'mokapi'

export default function() {
    app.websocket()
        .channel('/chat/{roomId}')
        .connect((event) => {
            const room = event.channel.params['roomId']
            console.log(`Client connected to room: ${room}`)
        })
        .message((event) => {
            const msg = event.message as { user: string; text: string }
            const room = event.channel.params['roomId']

            // Broadcast to all clients in this channel
            app.websocket().send(`/chat/${room}`, {
                user: msg.user,
                text: msg.text,
                timestamp: new Date().toISOString()
            })
        })
        .close((event) => {
            const room = event.channel.params['roomId']
            console.log(`Client disconnected from room: ${room}`)
        })

    app.websocket()
        .channel('/dm')
        .message((event) => {
            const msg = event.message as { to: string; text: string }
            // Reply only to the sender with an ack
            event.reply({ status: 'ok' })
        })

    app.websocket()
        .channel('/async')
        .message(async () => {
            const msg = { text: 'hello' }
            const sent = app.websocket().sendAsync('/async', msg)
            // the message must be sent as it was at call time
            msg.text = 'changed'
            await sent
        })
}
