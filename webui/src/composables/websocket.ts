export function useWebsocket() {

    function formatAddress(address: string): string {
        if (!address) {
            return address
        }
        return address.replace('[::1]', 'localhost')
    }

    function formatType(s: string): string {
        return s.charAt(0).toUpperCase() + s.slice(1)
    }

    return {
        formatAddress,
        formatType
    }
}