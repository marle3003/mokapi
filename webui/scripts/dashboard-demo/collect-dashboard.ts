import fs from 'fs/promises'

const baseUrl = 'http://localhost:8080';
const output = '../../public/demo'

export async function collectDashboard() {
  if (! await directoryExists(output)) {
    fs.mkdir(output)
  }

  const endpoints = {
    info: { path: '/api/info', loader: loadJson },
    services: { path: '/api/services' , loader: loadJson },
    metrics: { path: '/api/metrics?q=app', loader: loadJson },
    'service_Swagger Petstore': { path: '/api/services/http/Swagger%20Petstore', loader: loadJson },
    'service_Kafka Order Service API': { path: '/api/services/kafka/Kafka%20Order%20Service%20API', loader: loadJson },
    'service_Mail Server': { path: '/api/services/mail/Mail%20Server', loader: loadJson },
    'service_HR Employee Directory': { path: '/api/services/ldap/HR%20Employee%20Directory', loader: loadJson },
    events: { path: '/api/events', loader: fetchEvents },
    'mailbox_alice.johnson@example.com': { path: '/api/services/mail/Mail%20Server/mailboxes/alice.johnson@example.com', loader: loadJson },
    'mailbox_bob.miller@example.com': { path: '/api/services/mail/Mail%20Server/mailboxes/bob.miller@example.com', loader: loadJson },
    configs: { path: '/api/configs', loader: loadConfigs }
  }

  const snapshot: Record<string, any> = {}

  for (const [key, obj] of Object.entries(endpoints)) {
    const url = baseUrl + obj.path;
    await obj.loader(url, key, snapshot);
  }

  await fs.writeFile(output + '/dashboard.json', JSON.stringify(snapshot, null, 2));
}

async function loadJson(url: string, key: string, snapshot: Record<string, any>) {
  snapshot[key] = await fetchJson(url)
}

async function fetchJson(url: string): Promise<any> {
    try {
      const res = await fetch(url);
      if (!res.ok) {
        let text = await res.text()
        throw new Error(res.statusText + ': ' + text)
      }
      return await res.json()
    } catch(e) {
        console.error(`request ${url} failed: ${e}`)
    }
}

async function directoryExists(path: string) {
  try {
    const stats = await fs.stat(path);
    return stats.isDirectory();
  } catch (err: any) {
    return false
  }
}

async function fetchEvents(url: string, key: string, snapshot: Record<string, any>) {
  const events = await fetchJson(url);
  snapshot[key] = events

  const mails = [];
  for (const event of events) {
    if (event.traits.namespace !== 'mail') {
      continue;
    }

    const url = `${baseUrl}/api/services/mail/messages/${event.data.messageId}`
    const mail = await fetchJson(url);
    mails.push(mail)

    if (mail.data.attachments) {
      for (const attach of mail.data.attachments) {
        const data = await fetchBinary(`${baseUrl}/api/services/mail/messages/${event.data.messageId}/attachments/${attach.name}`);
        const filename = getFilenameWithRegex(attach.contentType);
        await fs.writeFile(`${output}/${filename}`, data);
      }
    }
  }
  snapshot['mails'] = mails
}

async function fetchBinary(url: string): Promise<any> {
    try {
      const res = await fetch(url);
      if (!res.ok) {
        let text = await res.text()
        throw new Error(res.statusText + ': ' + text)
      }
      const arrayBuffer = await res.arrayBuffer();
      return Buffer.from(arrayBuffer);
    } catch(e) {
        console.error(`request ${url} failed: ${e}`)
    }
}

function getFilenameWithRegex(contentType: string) {
    const match = contentType.match(/name=([^;]+)/);
    if (match && match[1]) {
        return match[1];
    }
    return null;
}

async function loadConfigs(url: string, key: string, snapshot:  Record<string, any>) {
    const configs = await fetchJson(url);
    snapshot[key] = configs;

    for (const config of configs) {
      const url = `${baseUrl}/api/configs/${config.id}`;
      const cfg = await fetchJson(url);
      snapshot['config_'+config.id] = cfg;
      const filename = getFilenameFromUrl(cfg.url);

      const data = await fetchBinary(`${baseUrl}/api/configs/${config.id}/data`);
      await fs.writeFile(`${output}/${filename}`, data);
    }
}

function getFilenameFromUrl(url: string): string {
  return new URL(url).pathname.split('/').pop()!;
}