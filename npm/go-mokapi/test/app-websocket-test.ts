import {
    app, WebsocketChannelRouter,
    WebsocketEventMessage, WebsocketEventConnect, WebsocketSendResult
} from "mokapi";

const channel: WebsocketChannelRouter = app.websocket().channel('foo')
channel.send({data: 123})

channel.connect((m: WebsocketEventConnect) => {
    // @ts-ignore
    m.message = 'foo';
    m.client.send('foo')
    m.reply({data: 123})
    m.broadcast({data: 123})
    m.channel.clients[0].send({})
})

channel.message((m: WebsocketEventMessage) => {
    // @ts-ignore
    m.message = 'foo';
    m.reply({data: 123})
})

app.websocket().message((m: WebsocketEventMessage) => {
    m.reply({data: 123})
})

// @ts-ignore
app.websocket().send({})
app.websocket().send('channel', {}, { retry: {factor: 10}, result: (r: WebsocketSendResult) => {r.api}})
app.websocket().api('title').send('channel', {})