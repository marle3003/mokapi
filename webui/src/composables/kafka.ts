export function useKafka() {

    function formatAddress(address: string): string {
        if (!address) {
            return address
        }
        return address.replace('[::1]', 'localhost')
    }

    return {
        formatAddress
    }
}