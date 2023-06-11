const defaultDescription = `Speed up testing process by creating stable development or test environments, reducing external dependencies, and simulating APIs that don't even exist yet.`

export function useMeta(title: string, description: string, canonicalUrl: string) {
    if (!description) {
        description = defaultDescription
    }
    canonicalUrl = canonicalUrl.replace('http://', 'https://')

    document.title = title
    setDescription(description)
    setCanonical(canonicalUrl)

    setOpenGraphMeta('og:url', canonicalUrl)
    setOpenGraphMeta('og:title', title)
    setOpenGraphMeta('og:description', description)
    setOpenGraphMeta('og:image', 'https://mokapi.io/og-logo.png')
    setOpenGraphMeta('og:image:alt', 'Mokapi logo')
    setOpenGraphMeta('og:type', 'website')
}

function setOpenGraphMeta(property: string, content: string) {
    let meta = document.head.querySelector(`meta[property="${property}"]`) as HTMLMetaElement
    if (!meta) {
        meta = document.createElement('meta')
        meta.setAttribute('property', property)
        document.getElementsByTagName('head')[0].prepend(meta);
    }
    meta.content = content;
}

function setDescription(description: string) {
    let meta = document.head.querySelector('meta[name="description"]') as HTMLMetaElement
    if (!meta) {
        meta = document.createElement('meta')
        meta.name = 'description'
        document.getElementsByTagName('head')[0].prepend(meta);
    }
    meta.content = description;
}

function setCanonical(href: string) {
    let canonical = document.head.querySelector('link[rel="canonical"]') as HTMLLinkElement
    if (!canonical) {
        canonical = document.createElement('link') as HTMLLinkElement;
        canonical.rel = 'canonical'
        document.getElementsByTagName('head')[0].appendChild(canonical);
    }
    canonical.href = href;
    
}