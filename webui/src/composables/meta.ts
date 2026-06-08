const defaultDescription = `Speed up testing process by creating stable development or test environments, reducing external dependencies, and simulating APIs that don't even exist yet.`

export function useMeta(title: string, description: string, canonicalUrl: string, image?: { url: string, alt: string}) {
    if (!description) {
        description = defaultDescription
    }
    canonicalUrl = canonicalUrl.replace('http://', 'https://')

    document.title = title
    setDescription(description)
    if (canonicalUrl) {
        setCanonical(canonicalUrl)
    }

    if (!image) {
        image = { url: '/og-logo.png', alt: 'Mokapi logo' }
    }

    setOpenGraphMeta('og:url', canonicalUrl)
    setOpenGraphMeta('og:title', title)
    setOpenGraphMeta('og:description', description)
    setOpenGraphMeta('og:image', 'https://mokapi.io' + image.url)
    setOpenGraphMeta('og:image:alt', image.alt)
    setOpenGraphMeta('og:type', 'website')
}

function setOpenGraphMeta(property: string, content: string) {
    let meta = document.head.querySelector(`meta[property="${property}"]`) as HTMLMetaElement
    if (!meta) {
        meta = document.createElement('meta')
        meta.setAttribute('property', property)
        document.getElementsByTagName('head')[0]!.prepend(meta);
    }
    meta.content = content;
}

function setDescription(description: string) {
    let meta = document.head.querySelector('meta[name="description"]') as HTMLMetaElement
    if (!meta) {
        meta = document.createElement('meta')
        meta.name = 'description'
        document.getElementsByTagName('head')[0]!.prepend(meta);
    }
    meta.content = description;
}

function setCanonical(href: string) {
    let canonical = document.head.querySelector('link[rel="canonical"]') as HTMLLinkElement
    if (!canonical) {
        canonical = document.createElement('link') as HTMLLinkElement;
        canonical.rel = 'canonical'
        document.getElementsByTagName('head')[0]!.appendChild(canonical);
    }
    canonical.href = href;
    
}

export function useSoftwareApplicationMeta(version: string) {
    var script = document.createElement('script');
    script.type = 'application/ld+json';
    script.text = JSON.stringify({
        "@context": "https://schema.org",
        "@type": "SoftwareApplication",
        "name": "Mokapi",
        "url": "https://mokapi.io",
        "applicationCategory": "DeveloperApplication",
        "operatingSystem": "Windows, macOS, Linux",
        "description": "An open-source, local-first multi-protocol mock API tool driven by OpenAPI and AsyncAPI specifications to simulate HTTP, Kafka, and other protocols.",
        "offers": {
            "@type": "Offer",
            "price": "0",
            "priceCurrency": "USD"
        },
        "downloadUrl": "https://github.com/marle3003/mokapi/releases",
        "softwareVersion": version, 
        "license": "https://github.com/marle3003/mokapi/blob/main/LICENSE",
        "features": "HTTP Mocking, Kafka Mocking, OpenAPI Support, AsyncAPI Support, Local-first API Simulation"
    }) 

    document.head.appendChild(script);
}