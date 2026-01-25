export function useKafka() {
    
    function clientSoftware(member: KafkaMember) {
        let client = `${member.clientSoftwareName} ${member.clientSoftwareVersion}`
        if (client === ' ') {
            client = '-'
        }
        return client
    }

    return {
        clientSoftware
    }
}