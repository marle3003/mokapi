const fs = require('fs');
const path = require('path');
const util = require('util');

const pages = [
  {
    file: './src/views/Home.vue',
    url: 'https://mokapi.io',
    priority: '1.0',
    changefreq: 'daily'
  },
  {
    file: './src/views/Http.vue',
    url: 'https://mokapi.io/http',
    priority: '0.8',
    changefreq: 'weekly'
  },
  {
    file: './src/views/Kafka.vue',
    url: 'https://mokapi.io/kafka',
    priority: '0.8',
    changefreq: 'weekly'
  },
  {
    file: './src/views/Mail.vue',
    url: 'https://mokapi.io/mail',
    priority: '0.8',
    changefreq: 'weekly'
  },
  {
    file: './src/views/Ldap.vue',
    url: 'https://mokapi.io/ldap',
    priority: '0.8',
    changefreq: 'weekly'
  },
]

const xmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
%s
</urlset>
`

const urlTemplate = `
<url>
  <loc>%s</loc>
  <changefreq>%s</changefreq>
  <priority>%s</priority>
  <lastmod>%s</lastmod>
</url>`

function writeItem(item) {
  let xml = '';

  if (item.path && item.source) {
    let changefreq = 'daily'
    let priority = '1.0'

    if (item.path.startsWith('/docs/')) {
      changefreq = 'weekly'
      priority = '0.6'
    }
    else if (item.path.startsWith('/resources/')) {
      changefreq = 'monthly'
      priority = '0.7'
    }

    stats = fs.statSync(path.join(docsPath, item.source));
    const url = `https://mokapi.io${item.path}`;
    const node = util.format(urlTemplate, url, changefreq, priority, stats.mtime.toISOString())
    xml += node
  }

  if (item.items) {
    for (const child of item.items) {
      xml += writeItem(child)
    }
  }
  return xml
}

function writeObject(obj, base) {

  let xml = ''
  for (let key in obj) {
    const item = obj[key];

    if (item.path && item.source) {
      stats = fs.statSync(path.join(docsPath, item.source));
      const url = `https://mokapi.io${item.path}`;
      const node = util.format(urlTemplate, url, '0.7', stats.mtime.toISOString())
      xml += node
    }

    if (item.items) {
      for (const child of item.items) {
        xml += writeItem(child)
      }
    }
  }
  return xml
}

const docsPath = '../docs'

try {
  let content = ''

  // write pages
  for (let page of pages) {
    const stats = fs.statSync(page.file)
    content += util.format(urlTemplate, page.url, page.changefreq, page.priority, stats.mtime.toISOString())
  }

  // write docs
  const data = fs.readFileSync(path.join(docsPath, 'config.json'), 'utf8');
  docs = JSON.parse(data)
  content += writeObject(docs, '')

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