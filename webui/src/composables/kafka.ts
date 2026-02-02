export function useKafka() {
    
    function clientSoftware(member: KafkaClient | KafkaMember) {
        let client = `${member.clientSoftwareName} ${member.clientSoftwareVersion}`
        if (client === ' ') {
            client = '-'
        }
        return client
    }

    function formatAddress(address: string): string {
        return address.replace('[::1]', 'localhost')
    }

    return {
        clientSoftware,
        formatAddress
    }
}