const defaultDescription = `Speed up testing process by creating stable development or test environments, reducing external dependencies, and simulating APIs that don't even exist yet.`

// A tag of the document head, identified by a CSS selector. The browser applies it to `document`
// (updating the tag of the pre-rendered page if there is one), on the server it is collected per
// page, see ssg/prerender.js.
export type HeadTag = {
    selector: string
    tagName: 'meta' | 'link' | 'script'
    attributes: Record<string, string>
    text?: string
}

let collectedTitle: string | undefined
const collectedTags = new Map<string, HeadTag>()

export function resetCollectedHead() {
    collectedTitle = undefined
    collectedTags.clear()
}

export function collectedHead(): { title: string | undefined, tags: HeadTag[] } {
    return { title: collectedTitle, tags: [...collectedTags.values()] }
}

export function setTitle(title: string) {
    if (import.meta.env.SSR) {
        collectedTitle = title
    } else {
        document.title = title
    }
}

export function setHeadTag(tag: HeadTag) {
    if (import.meta.env.SSR) {
        collectedTags.set(tag.selector, tag)
        return
    }

    let element = document.head.querySelector(tag.selector)
    if (!element) {
        element = document.createElement(tag.tagName)
        document.head.appendChild(element)
    }
    for (const [name, value] of Object.entries(tag.attributes)) {
        element.setAttribute(name, value)
    }
    if (tag.text !== undefined) {
        element.textContent = tag.text
    }
}

export function removeHeadTag(selector: string) {
    if (import.meta.env.SSR) {
        collectedTags.delete(selector)
    } else {
        document.head.querySelector(selector)?.remove()
    }
}

function metaTag(key: 'name' | 'property', name: string, content: string): HeadTag {
    return { selector: `meta[${key}="${name}"]`, tagName: 'meta', attributes: { [key]: name, content } }
}

export function useMeta(title: string, description: string, canonicalUrl: string, image?: { url: string, alt: string}) {
    if (!description) {
        description = defaultDescription
    }
    canonicalUrl = canonicalUrl.replace('http://', 'https://')

    setTitle(title)
    setHeadTag(metaTag('name', 'description', description))
    if (canonicalUrl) {
        setHeadTag({ selector: 'link[rel="canonical"]', tagName: 'link', attributes: { rel: 'canonical', href: canonicalUrl } })
    }

    if (!image) {
        image = { url: '/og-logo.png', alt: 'Mokapi logo' }
    }

    setHeadTag(metaTag('property', 'og:url', canonicalUrl))
    setHeadTag(metaTag('property', 'og:title', title))
    setHeadTag(metaTag('property', 'og:description', description))
    setHeadTag(metaTag('property', 'og:image', 'https://mokapi.io' + image.url))
    setHeadTag(metaTag('property', 'og:image:alt', image.alt))
    setHeadTag(metaTag('property', 'og:type', 'website'))
}

export function useSoftwareApplicationMeta(version: string) {
    const structuredData = {
        "@context": "https://schema.org",
        "@type": "SoftwareApplication",
        "name": "Mokapi",
        "url": "https://mokapi.io",
        "sameAs": [
            "https://github.com/marle3003/mokapi",
            "https://hub.docker.com/r/mokapi/mokapi",
            "https://www.npmjs.com/package/go-mokapi"
        ],
        "applicationCategory": "DeveloperApplication",
        "operatingSystem": "Windows, macOS, Linux",
        "description": "An open-source, local-first multi-protocol mock API tool driven by OpenAPI and AsyncAPI specifications to simulate HTTP, Kafka, WebSocket, MQTT, SMTP, and LDAP.",
        "offers": {
            "@type": "Offer",
            "price": "0",
            "priceCurrency": "USD"
        },
        "softwareRequirements": "Docker, Go runtime, or Node.js/npm",
        "downloadUrl": "https://github.com/marle3003/mokapi/releases",
        "softwareVersion": version, 
        "license": "https://github.com/marle3003/mokapi/blob/main/LICENSE",
        "keywords": "Mokapi, API simulation engine, Service Virtualization, Apache Kafka mock server, local LDAP mock, AsyncAPI simulation, offline development container, Go software",
        "featureList": [
            "Multi-Protocol Service Virtualization",
            "Apache Kafka Event Stream Simulation",
            "Local LDAP and SMTP Testing Server",
            "Contract Testing via OpenAPI and AsyncAPI",
            "JavaScript-driven Mock Behavior Scripting"
        ],
        "author": {
            "@type": "Person",
            "name": "Marcel Lehmann",
            "url": "https://github.com/marle3003"
        },
    }

    setHeadTag({
        selector: 'script[type="application/ld+json"]',
        tagName: 'script',
        attributes: { type: 'application/ld+json' },
        text: JSON.stringify(structuredData),
    })
}