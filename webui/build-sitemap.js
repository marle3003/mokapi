const fs = require('fs');
const path = require('path');
const util = require('util');

const pages = [
  {
    file: './src/views/Home.vue',
    url: 'https://mokapi.io',
  },
  {
    file: './src/views/Http.vue',
    url: 'https://mokapi.io/http',
  },
  {
    file: './src/views/Kafka.vue',
    url: 'https://mokapi.io/kafka',
  },
  {
    file: './src/views/Mqtt.vue',
    url: 'https://mokapi.io/mqtt',
  },
  {
    file: './src/views/Mail.vue',
    url: 'https://mokapi.io/mail',
  },
  {
    file: './src/views/Ldap.vue',
    url: 'https://mokapi.io/ldap',
  },
  {
    file: './src/views/Compare.vue',
    url: 'https://mokapi.io/compare',
  },
]

const xmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
%s
</urlset>
`

const urlTemplate = `<url>
  <loc>%s</loc>
  <lastmod>%s</lastmod>
</url>`

function writeItem(item) {
  let xml = '';

  if (item.path && item.source) {
    stats = fs.statSync(path.join(docsPath, item.source));
    const url = `https://mokapi.io${item.path}`;
    const node = util.format(urlTemplate, url, stats.mtime.toISOString())
    xml += node + '\n'
  }

  if (item.items) {
    for (const child of item.items) {
      xml += writeItem(child)
    }
  }
  return xml
}

function writeObject(obj) {
  let xml = ''
  for (let key in obj) {
    const item = obj[key];
    xml += writeItem(item)
  }
  return xml
}

const docsPath = '../docs'

try {
  let content = ''

  // write pages
  for (let page of pages) {
    const stats = fs.statSync(page.file)
    content += util.format(urlTemplate, page.url, stats.mtime.toISOString())
  }

  // write docs
  const data = fs.readFileSync(path.join(docsPath, 'config.json'), 'utf8');
  docs = JSON.parse(data)
  content += writeObject(docs)

  // trim leading and trailing whitespace and new lines
  content = content.replace(/^\s+|\s+$/g, '');

  // write to file
  const xml = util.format(xmlTemplate, content)
  fs.writeFileSync('./public/sitemap.xml', xml, { flag: 'w' })
  if (!fs.existsSync('dist')) {
    fs.mkdirSync('dist')
  }
  fs.writeFileSync('./dist/sitemap.xml', xml, { flag: 'w' })
} catch (err) {
  console.error(err);
}