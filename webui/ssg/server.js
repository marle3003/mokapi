const http = require('http');
const fs = require('fs');
const path = require('path');
const mime = require('mime-types');

module.exports = class Server {
    server

    constructor(baseDir) {
        this.server = http.Server(async function (request, response) {
            var filePath = request.url.split('?')[0]
            if (filePath == '') {
                filePath = 'index.html'
            }

            var contentType = 'text/html';
            if (await fileExists(path.join(baseDir, filePath))) {
                var extname = path.extname(filePath);
                if (extname !== '') {
                    contentType = mime.lookup(filePath)
                } else {
                    filePath = 'index.html'
                }
            } else {
                filePath = 'index.html'
            }

            filePath = path.join(baseDir, filePath)
            fs.readFile(filePath, function(error, content) {
                if (error) {
                    console.error('ERROR: '+error)
                    if(error.code == 'ENOENT'){
                        fs.readFile('./404.html', function(error, content) {
                            response.writeHead(200, { 'Content-Type': contentType });
                            response.end(content, 'utf-8');
                        });
                    }
                    else {
                        response.writeHead(500);
                        response.end('Sorry, check with the site admin for error: '+error.code+' ..\n');
                        response.end(); 
                    }
                }
                else {
                    response.writeHead(200, { 'Content-Type': contentType });
                    response.end(content, 'utf-8');
                }
            });
        })
    }

    async start() {
        this.server.listen(8025)
    }

    close() {
        this.server.close()
    }
}

async function fileExists(path) {
    try {
        const stats = await fs.promises.stat(path);
        return stats.isFile();
    } catch (err) {
        return false
    }
}