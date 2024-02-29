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
    if (visited.has(url.pathname)) {
      return
    }
    visited.add(url.pathname)
    console.info(`crawling ${url.href}...`)
    const page = await browser.newPage();
    await page.goto(url.href, {
      waitUntil: 'networkidle',
    });

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

    const p = path.join('../dist', url.pathname)
    if (!fs.existsSync(path.dirname(p))){
      fs.mkdirSync(path.dirname(p), { recursive: true });
    }

    let content = await page.content();
    content = content.replace(/http:\/\/localhost:8025/g, '')
    fs.writeFileSync(p + '.html', content);

    let links = new Set(await page.evaluate(async (url) => {
      return Array.from(document.querySelectorAll('a'))
        .map((a) => new URL(a.href))
        .filter((u) => u.hostname == url.hostname)
    }, url))
    await page.close()

    //const promises = [];
    for (const u of links) {
      //promises.push(crawl(u))
      await crawl(u)
    }
    //await Promise.all(promises)
  }

  await crawl(new URL('http://localhost:8025/home'))

  await browser.close();

  server.close()
})();