import { usePrettyLanguage } from '@/composables/usePrettyLanguage'

const binaryContentTypes = ['avro/binary', 'application/avro', 'application/octet-stream']

// A recorded message value shown the way its message configuration (content type, Avro) says.
// Used by the Kafka and MQTT message views.
export function messageValue(config: BaseMessage, message: { value?: string, binary?: string }) {
    const { formatLanguage } = usePrettyLanguage()
    const [contentType, isAvro] = getContentType(config)

    const source: Source = {}
    if (message.value) {
        source.preview = {
            content: formatLanguage(message.value, contentType),
            contentType,
            contentTypeTitle: config.contentType,
            description: isAvro ? 'Avro content in JSON format' : undefined,
        }
    }
    if (message.binary && isBinaryContentType(config.contentType)) {
        if (source.preview) {
            source.preview.description = 'Avro content in JSON format'
        }
        source.binary = { content: atob(message.binary), contentType: config.contentType }
    }
    return { source, contentType, contentTypeTitle: config.contentType }
}

function isBinaryContentType(contentType: string): boolean {
    return binaryContentTypes.includes(contentType)
}

// An Avro message is shown as JSON, its binary form is a second view of the same value.
function getContentType(config: BaseMessage): [string, boolean] {
    if (config.payload.format?.includes('application/vnd.apache.avro') && isBinaryContentType(config.contentType)) {
        return ['application/json', true]
    }
    return [config.contentType, false]
}
