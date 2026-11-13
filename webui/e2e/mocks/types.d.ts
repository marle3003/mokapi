interface Config {
    name: string;
    description?: string;
    version?: string;
    contact?: Contact;
    status?: string;
}

interface Info {
    type: string;
    name: string;
    description?: string;
    version?: string;
    contact?: Contact;
    status?: string;
}

interface Contact {
    name?: string;
    email?: string;
    url?: string;
}

interface Label {
    name: string;
    value: string;
}

type Traits = Record<string, string | undefined>;

interface ServiceEvent {
    id: string;
    traits: Traits;
    time: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- event data differs per namespace (http, kafka, mail, ...)
    data: any
}