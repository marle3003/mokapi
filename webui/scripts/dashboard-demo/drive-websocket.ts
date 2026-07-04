export async function driveWebsocket(): Promise<void> {

    const bob = new WebSocket('ws://localhost:8000/chats/general');
    const carol = new WebSocket('ws://localhost:8000/chats/general');

    carol.addEventListener('message', console.log)

    // Wait for connection to open
    await new Promise((resolve, reject) => {
        bob.addEventListener('open', resolve, { once: true });
        bob.addEventListener('error', reject, { once: true });
    });

    bob.addEventListener('error', error => {
        console.error('WebSocket error:', error);
    });

    return safeSend(bob, JSON.stringify({ text: 'Hello there!', username: 'bob'}))
}

function safeSend(socket: WebSocket, message: string, waitMs = 100) {
  return new Promise<void>((resolve, reject) => {
    const onClose = (event: CloseEvent) => {
      cleanup();
      reject(new Error(`Rejected — code: ${event.code}, reason: ${event.reason || 'none'}`));
    };
    const onError = () => {
      cleanup();
      reject(new Error('WebSocket error'));
    };

    const cleanup = () => {
      socket.removeEventListener('close', onClose);
      socket.removeEventListener('error', onError);
    };

    socket.addEventListener('close', onClose, { once: true });
    socket.addEventListener('error', onError, { once: true });

    socket.send(message);

    // If connection is still alive after waitMs, assume send was accepted
    setTimeout(() => {
      cleanup();
      resolve();
    }, waitMs);
  });
}