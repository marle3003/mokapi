// Structural types: helpers.ts is also loaded by Vitest under the webui tsconfig, where the globals
// `Config` and `Info` of e2e/mocks/types.d.ts don't exist (src/types has a different `Config`).
interface Contact {
    name?: string
    email?: string
    url?: string
}

interface ConfigInfo {
    name: string
    description?: string
    version?: string
    contact?: Contact
    status?: string
}

type ServiceInfo = ConfigInfo & { type: string }

export function getInfo(config: ConfigInfo, type: string) {
    const info: ServiceInfo = {
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

export function matchTraits(e: { traits: Record<string, string | undefined> }, traits: Record<string, string | undefined>) {
    for (const t in traits) {
        if (e.traits[t] !== traits[t]) {
            return false
        }
    }
    return true
}