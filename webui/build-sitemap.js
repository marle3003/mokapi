const fs = require('fs');
const util = require('util');

const xmlTemplate = `
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:news="http://www.google.com/schemas/sitemap-news/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1" xmlns:video="http://www.google.com/schemas/sitemap-video/1.1">
<url>
  <loc>https://mokapi.io/home/</loc>
  <changefreq>daily</changefreq>
  <priority>0.8</priority>
</url>
%s
</urlset>
`

const urlTemplate = `
<url>
  <loc>%s/</loc>
  <changefreq>daily</changefreq>
  <priority>0.8</priority>
  <lastmod>%s</lastmod>
</url>`

lastModified = new Date().toISOString()

function writeObject(obj, base) {
    let xml = ''
    for (k in obj) {
        let segment = k.split(' ').join('-').split('/').join('-').replace('&', '%26')
        const path = base + '/' + segment.toLowerCase()

        if (typeof obj[k] !== "string") {
            xml += writeObject(obj[k], path)
        }else{
            const url = 'https://mokapi.io/docs' + path
            const node = util.format(urlTemplate, url, lastModified)
            xml += node
        }
    }
    return xml
}

try {
    const data = fs.readFileSync('./src/assets/docs/config.json', 'utf8');
    docs = JSON.parse(data)
    let xml = writeObject(docs, '')
    xml = util.format(xmlTemplate, xml)
    fs.writeFileSync('./public/sitemap.xml', xml, { flag: 'w' })
  } catch (err) {
    console.error(err);
  }