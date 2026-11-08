import { app, WebsocketChannelRouter, WebsocketEventConnect, WebsocketEventMessage, WebsocketRouter, WebsocketSendResult } from "mokapi";

const channel: WebsocketChannelRouter = app.websocket().channel("foo");
channel.send({ data: 123 });

channel.connect((m: WebsocketEventConnect) => {
    // @ts-expect-error
    m.message = "foo";
    m.client.send("foo");
    m.reply({ data: 123 });
    m.broadcast({ data: 123 });
    m.channel.clients[0].send({});
});

channel.message((m: WebsocketEventMessage) => {
    // @ts-expect-error
    m.message = "foo";
    m.reply({ data: 123 });
});

app.websocket().message((m: WebsocketEventMessage) => {
    m.reply({ data: 123 });
});

// @ts-expect-error
app.websocket().send({});
app.websocket().send("channel", {}, {
    retry: { factor: 10 },
    result: (r: WebsocketSendResult) => {
        r.api;
    },
});
app.websocket().api("title").send("channel", {});

async function sendAsync() {
    const router: WebsocketRouter = await app.websocket().sendAsync("channel", {}, {
        retry: { factor: 10 },
        result: (r: WebsocketSendResult) => {
            r.api;
        },
    });
    const channelRouter: WebsocketChannelRouter = await app.websocket().channel("channel").sendAsync({});
    // @ts-expect-error
    await app.websocket().sendAsync({});
}

const routerAfterSend: WebsocketRouter = app.websocket().send("channel", {});
const channelRouterAfterSend: WebsocketChannelRouter = app.websocket().channel("channel").send({});

app.websocket().send("channel", undefined, { messageId: "serverMessage" });
app.websocket().send("channel", { text: "hello" }, { messageId: "serverMessage" });
app.websocket().channel("channel").send(undefined, { messageId: "serverMessage" });
app.websocket().sendAsync("channel", undefined, { messageId: "serverMessage" });
app.websocket().channel("channel").sendAsync(undefined, { messageId: "serverMessage" });
// @ts-expect-error
app.websocket().send("channel", undefined, { messageId: 1 });
