export function useMqtt() {

    function fromatVersion(version: number) {
        switch (version) {
            case 5:
                return `${version} (v5)`
            case 4:
                return `${version} (v3.1.1)`
            case 3:
                return '${version} (v3.1)';
            default:
                return `${version} (Unknown)`;
        }
    }

    function formatAddress(address: string): string {
        if (!address) {
            return address
        }
        return address.replace('[::1]', 'localhost')
    }

    function formatType(type: number) {
        switch (type) {
            case 1: return 'Connect'
            case 8: return 'Subscribe'
            case 14: return 'Disconnect'
            default: return `unknown (${type})`
        }
    }

    function formatQoS(qos: number) {
        switch (qos) {
            case 0: return '0 (at most once)'
            case 1: return '1 (at least once)'
            case 2: return '2 (exactly once)'
        }
    }

    function formatDisconnectReason(reason: number) {
        switch (reason) {
            case 0: return '0 (Normal disconnection)'
            case 4: return '4 (Disconnect with Will Message)'

            case 128: return '128 (Unspecified error)'
            case 129: return '129 (Malformed Packet)'
            case 130: return '130 (Protocol Error)'
            case 131: return '131 (Implementation specific error)'
            case 135: return '135 (Not authorized)'
            case 137: return '137 (Server busy)'
            case 139: return '139 (Server shutting down)'
            case 141: return '141 (Keep Alive timeout)'
            case 142: return '142 (Session taken over)'
            case 143: return '143 (Topic Filter invalid)'
            case 144: return '144 (Topic Name invalid)'
            case 147: return '147 (Receive Maximum exceeded)'
            case 148: return '148 (Topic Alias invalid)'
            case 149: return '149 (Packet too large)'
            case 150: return '150 (Message rate too high)'
            case 151: return '151 (Quota exceeded)'
            case 152: return '152 (Administrative action)'
            case 153: return '153 (Payload format invalid)'
            case 154: return '154 (Retain not supported)'
            case 155: return '155 (QoS not supported)'
            case 156: return '156 (Use another server)'
            case 157: return '157 (Server moved)'
            case 158: return '158 (Shared Subscriptions not supported)'
            case 159: return '159 (Connection rate exceeded)'
            case 160: return '160 (Maximum connect time)'
            case 161: return '161 (Subscription Identifiers not supported)'
            case 162: return '162 (Wildcard Subscriptions not supported)'
            default: return `${reason} (unknown)`
        }
    }

    return {
        fromatVersion,
        formatAddress,
        formatType,
        formatQoS,
        formatDisconnectReason
    }
}