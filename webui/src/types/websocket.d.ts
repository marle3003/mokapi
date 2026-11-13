declare interface WebsocketService extends BaseChannelService<WebsocketChannel> {
  servers: WebsocketServer[];
  clients?: WebsocketClient[];
}

declare interface WebsocketServer {
  name: string;
  host: string;
  protocol: string
  title: string
  summary: string
  description: string;
  tags: Tag[]
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- arbitrary JSON from the dashboard API
  configs: { [key: string]: any }
}

declare interface WebsocketChannel extends BaseChannel {
  metrics: {
    messages_total: number
    message_timestamp: number
  }
}

declare interface WebsocketServer {
  name: string;
  addr: string;
}

declare type WebsocketEventInfo =  WebsocketMessageInfo | WebsocketConnectionLog

declare interface WebsocketMessageInfo {
  channel: string
  message: WebsocketMessage;
  direction: string
  client: WebsocketClientLog
  validationErrors?: Record<string, string>
}

declare type WebsocketEventData = WebsocketMessageData | WebsocketConnectionLog

declare interface WebsocketMessageData {
  api: string
  channel: string
  message: WebsocketMessage;
  messageId: string
  direction: 'receive' | 'send'
  client: WebsocketClientLog
  actions: Action[]
  validationErrors?: Record<string, string>
}

declare interface WebsocketMessage extends BaseMessage {
  value?: string
  binary?: string
}

declare interface WebsocketClient {
  id: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- arbitrary JSON from the dashboard API
  query: Record<string, any>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- arbitrary JSON from the dashboard API
  headers: Record<string, any>
  address: string
  serverAddress: string
}

declare interface WebsocketClientLog {
  id: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- arbitrary JSON from the dashboard API
  query: Record<string, any>
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- arbitrary JSON from the dashboard API
  headers: Record<string, any>
  address: string
  server: string
}

declare interface WebsocketConnectionLog {
  type: 'connect' | 'close'
  api: string
  channel: string
  client: WebsocketClientLog
  actions: Action[]
}

declare interface WebsocketCloseLog extends WebsocketConnectionLog {
  reason: string
  closedBy: string
}