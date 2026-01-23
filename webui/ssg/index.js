const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');
const Server = require('./server');

(async () => {
  const server = new Server('../dist')
  await server.start()

  const visited = new Set();
  visited.add('/')
  const browser = await chromium.launch();

  async function crawl(url) {
    if (!url.href.startsWith('http://localhost:8025') || url.href === 'http://localhost:8025/dashboard') {
      console.log('skip ' + url)
      return
    }
    if (path.extname(url.pathname) === '.ts') {
      console.log('skip ' + url);
      return;
    }
    if (visited.has(url.pathname)) {
      return
    }
    visited.add(url.pathname)
    console.info(`crawling ${url.href}...`)
    const page = await browser.newPage();
    const response = await page.goto(url.href, {
      waitUntil: 'networkidle',
    });

    if (!response) {
      throw new Error(`page ${url.href} not found`);
    }

    const contentType = response.headers()['content-type'] ?? '';
    if (!contentType.includes('text/html')) {
      console.log('skip ' + contentType);
      return
    }

    await page.evaluate(async () => {
      for (const script of document.querySelectorAll('script')) {
        script.remove();
       }
       for (const link of document.querySelectorAll('link'))  {
           if (link.as === 'script') {
            link.remove();
           }
       }
    });

    const [pageFound, msg] = await page.evaluate(async () => {
      const header = document.querySelector('h1')
      if (!header) {
        return [false, `header <h1> not found`]
      }
      if (header.innerText === `Sorry, this page isn't available`) {
        return [false, `Sorry, this page isn't available`]
      }
      return [true, '']
    })
    if (!pageFound) {
      throw new Error(`page ${url.href} not found: ${msg}`)
    }

    let p = path.join('../dist', decodeURIComponent(url.pathname))
    if (!fs.existsSync(path.dirname(p))){
      fs.mkdirSync(path.dirname(p), { recursive: true });
    }

    let content = await page.content();
    content = content.replace(/http:\/\/localhost:8025/g, '')

    let links = new Set(await page.evaluate(async () => {
      return Array.from(document.querySelectorAll('a'))
        .map((a) => a.href)
    }))
    await page.close()

    links = links.keys()
      .map(s => {
        const u = URL.parse(s)
        if (u === null) {
          console.log('warning: invalid URL: ' + s)
        }
        return u
      })
      .filter((u) => u !== null && u.hostname == url.hostname)

    for (const u of links) {
      try {
        await crawl(u)
      } catch (err) {
        if (err.message && err.message.startsWith('page ')) { 
          throw new Error(`crawl link on page ${url} failed: ${err}`)
        }
        throw err
      }
    }

    if (isDir(p)) {
      p += '/index'
    }

    fs.writeFileSync(p + '.html', content);
  }

  await crawl(new URL('http://localhost:8025/home'))

  await browser.close();

  server.close()
})();

function isDir(path) {
  try {
      const stats = fs.statSync(path);
      return stats.isDirectory();
  } catch (err) {
      return false;
  }
}