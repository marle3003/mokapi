export function getInfo(config: Config, type: string) {
    const info: Info = {
        name: config.name,
        description: config.description,
        version: config.version,
        type: type,
    }
    if (config.contact) {
        info.contact = {
            name: config.contact.name,
            email: config.contact.email,
            url: config.contact.url
        }
    }
    if (config.status) {
        info.status = config.status
    }
    return info
}

export function matchTraits(e: ServiceEvent, traits: Traits) {
    for (let t in traits) {
        if (e.traits[t] !== traits[t]) {
            return false
        }
    }
    return true
}