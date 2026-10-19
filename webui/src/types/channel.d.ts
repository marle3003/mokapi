declare interface BaseChannelInfo {
  name: string
  summary?: string
  parameters?: unknown[]
  instances?: ChannelInstance[]
  tags: Tag[]
}

declare interface BaseChannel<TMessage extends BaseMessage = BaseMessage> extends BaseChannelInfo {
  title?: string
  description?: string
  messages: { [messageId: string]: TMessage }
}

declare interface ChannelInstance {
  name: string
  parameters: Record<string, string>
  metrics: {
    messages_total: number
    message_timestamp: number
  }
}

declare interface BaseChannelService<Channel extends BaseChannelInfo = BaseChannelInfo> extends Service {
  name: string
  channels: Channel[]
}

declare interface BaseTopicService<Channel extends BaseChannelInfo = BaseChannelInfo> extends Service {
  name: string
  topics: Channel[]
}

declare interface BaseMessage {
  name: string
  title: string
  summary: string
  description: string
  payload: SchemaFormat;
  contentType: string;
}

declare interface Tag {
  name: string
  description: string
}