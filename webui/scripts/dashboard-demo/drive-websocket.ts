export async function driveWebsocket(): Promise<void> {

    const bob = new WebSocket('ws://localhost:8000/chats/general');
    const carol = new WebSocket('ws://localhost:8000/chats/general');

    // Mokapi broadcasts only to clients that completed the handshake
    await Promise.all([opened(bob), opened(carol)]);

    // Listen before sending, otherwise the broadcast could be missed
    const broadcastReceived = Promise.all([
        receivedMessage('bob', bob),
        receivedMessage('carol', carol),
    ]);

    bob.send(JSON.stringify({ text: 'Hello there!', username: 'bob'}));

    await broadcastReceived;
}

function opened(socket: WebSocket) {
    return new Promise<void>((resolve, reject) => {
        socket.addEventListener('open', () => resolve(), { once: true });
        socket.addEventListener('error', () => reject(new Error('WebSocket error before open')), { once: true });
    });
}

function receivedMessage(name: string, socket: WebSocket, timeoutMs = 5000) {
    return new Promise<void>((resolve, reject) => {
        const fail = (reason: string) => {
            cleanup();
            reject(new Error(`websocket: ${name} did not receive the broadcast (${reason})`));
        };
        const onMessage = (event: MessageEvent) => {
            cleanup();
            console.log(`websocket: ${name} received`, event.data);
            resolve();
        };
        const onClose = (event: CloseEvent) => fail(`closed, code: ${event.code}, reason: ${event.reason || 'none'}`);
        const onError = () => fail('WebSocket error');
        const timer = setTimeout(() => fail(`timeout after ${timeoutMs} ms`), timeoutMs);

        const cleanup = () => {
            clearTimeout(timer);
            socket.removeEventListener('message', onMessage);
            socket.removeEventListener('close', onClose);
            socket.removeEventListener('error', onError);
        };

        socket.addEventListener('message', onMessage);
        socket.addEventListener('close', onClose);
        socket.addEventListener('error', onError);
    });
}
