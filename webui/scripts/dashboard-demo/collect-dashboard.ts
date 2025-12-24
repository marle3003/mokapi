import fs from 'fs/promises'

const baseUrl = 'http://localhost:8080';
const output = '../../public/demo'

export async function collectDashboard() {
  if (! await directoryExists(output)) {
    fs.mkdir(output)
  }

  const endpoints = {
    services: '/api/services',
    metrics: '/api/metrics?q=app',
    'service_Swagger Petstore': '/api/services/http/Swagger%20Petstore',
    'service_Kafka Order Service API': '/api/services/kafka/Kafka%20Order%20Service%20API',
    'service_Mail Server': '/api/services/mail/Mail%20Server',
    'service_HR Employee Directory': '/api/services/ldap/HR%20Employee%20Directory',
    events: '/api/events',
    'mailbox_alice.johnson@example.com': '/api/services/mail/Mail%20Server/mailboxes/alice.johnson@example.com',
    'mailbox_bob.miller@example.com': '/api/services/mail/Mail%20Server/mailboxes/bob.miller@example.com',
  }

  const snapshot: Record<string, any> = {}

  for (const [key, path] of Object.entries(endpoints)) {
    const url = baseUrl + path;
    if (key === 'events') {
      await fetchEvents(snapshot, url);
    } else {
      snapshot[key] = await fetchJson(url);
    }
  }

  await fs.writeFile(output + '/dashboard.json', JSON.stringify(snapshot, null, 2));
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

async function fetchEvents(snapshot: Record<string, any>, url: string) {
  const events = await fetchJson(url);
  snapshot['events'] = events;

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