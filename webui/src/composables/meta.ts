
export function useMeta(title: string, description: string) {
    document.title = title
    setMeta('description', description)

    setOpenGraphMeta('og:site_name', 'Mokapi')
    setOpenGraphMeta('og.url', "https://mokapi.io")
    setOpenGraphMeta('og:title', title)
    setOpenGraphMeta('og:description', description)
    setOpenGraphMeta('og:image', 'https://mokapi.io/og-logo.png')
    setOpenGraphMeta('og:image:alt', 'Mokapi logo')
    setOpenGraphMeta('og:type', 'website')
}

function setOpenGraphMeta(property: string, content: string) {
    var meta = document.createElement('meta');
    meta.setAttribute('property', property)
    meta.content = content;
    document.getElementsByTagName('head')[0].prepend(meta);
}

function setMeta(name: string, content: string) {
    var meta = document.createElement('meta') as HTMLMetaElement;
    meta.name = name
    meta.content = content;
    document.getElementsByTagName('head')[0].appendChild(meta);
}