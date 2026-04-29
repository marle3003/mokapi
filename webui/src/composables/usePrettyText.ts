
export function usePrettyText() {

    function parseUrls(input: string) {
        const urlRegex = /(\b(?:https?:\/\/|www\.)[^\s<]+[^\s<.,:;"')\]\s])/gi
        return input.replace(urlRegex, url => {
            return `<a href="${url}" rel="noopener">${url}</a>`
        })
    }

    function adaptiveTruncate(s: string, maxLength = 40) {
        if (s.length <= maxLength) {
            return s;
        }

        const segments = s.split('/').filter(Boolean);
        const isAbsolute = s.startsWith('/');

        if (segments.length > 2) {
            const start = `${isAbsolute ? '/' : ''}${segments[0]}/.../`
            let end = segments[segments.length - 1]
            let n = maxLength - start.length - end.length
            for (let i = segments.length - 2; i > 0; i--) {
                n -= segments[i].length + 1
                if (n < 0) {
                    break
                }
                end = `${segments[i]}/${end}`
            }
            return `${start}${end}`;
        }

        if (s.length > maxLength) {
            const charsToKeep = Math.floor((maxLength - 3) / 2);
            return `${s.substring(0, charsToKeep)}...${s.slice(-charsToKeep)}`;
        }

        return s;
    }

    function fromBinary(encoded: string) {
        const binary = atob(encoded);
        const bytes = new Uint8Array(binary.length);
        for (let i = 0; i < bytes.length; i++) {
            bytes[i] = binary.charCodeAt(i);
        }
        return String.fromCharCode(...new Uint16Array(bytes.buffer));
    }

    return { parseUrls, adaptiveTruncate, fromBinary }
}