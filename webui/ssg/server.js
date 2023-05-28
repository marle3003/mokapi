var http = require('http')
var fs = require('fs')
var path = require('path')

module.exports = class Server {
    server

    constructor(baseDir) {
        this.server = http.Server(function (request, response) {
            var filePath = request.url.split('?')[0]
            if (filePath == '') {
                filePath = 'index.html'
            }

            var extname = path.extname(filePath)
            var contentType = 'text/html'
            switch (extname) {
                case '.js':
                    contentType = 'text/javascript'
                    break
                case '.css':
                    contentType = 'text/css'
                    break
                case '.json':
                    contentType = 'application/json'
                    break
                case '.png':
                    contentType = 'image/png'
                    break
                case '.jpg':
                    contentType = 'image/jpg';
                    break
                case '.wav':
                    contentType = 'audio/wav'
                    break
                case '':
                    filePath = 'index.html'
            }

            filePath = path.join(baseDir, filePath)
            fs.readFile(filePath, function(error, content) {
                if (error) {
                    console.log('ERROR: '+error)
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