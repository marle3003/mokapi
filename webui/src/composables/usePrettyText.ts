
export function usePrettyText() {

    function parseUrls(input: string) {
        const urlRegex = /(\b(?:https?:\/\/|www\.)[^\s<]+[^\s<.,:;"')\]\s])/gi
        return input.replace(urlRegex, url => {
            return `<a href="${url}" rel="noopener">${url}</a>`
        })
    }

    return { parseUrls }
}