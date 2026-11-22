export function formatAddress(address: string): string {
    if (!address) {
        return address
    }
    return address.replace('[::1]', 'localhost')
}

export function formatType(s: string): string {
    return s.charAt(0).toUpperCase() + s.slice(1)
}

// `send` is a message the client sent to the server, `receive` one the server sent to the client.
export function messageSource(data: WebsocketMessageData): string {
    return formatAddress(data.direction === 'send' ? data.client.address : data.client.server)
}

export function messageDestination(data: WebsocketMessageData): string {
    return formatAddress(data.direction === 'receive' ? data.client.address : data.client.server)
}

// The channel a recorded path belongs to: the path of a template channel is one of its instances.
export function findChannel(service: WebsocketService, path: string | undefined): WebsocketChannel | undefined {
    return service.channels.find(c => c.name === path || c.instances?.some(i => i.name === path))
}

// The client view exists as long as the service still lists the client.
export function hasClient(service: WebsocketService, clientId: string): boolean {
    return service.clients?.some(c => c.id === clientId) ?? false
}

export function useWebsocket() {
    return {
        formatAddress,
        formatType
    }
}
