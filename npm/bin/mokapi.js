#!/usr/bin/env node
let spawn = require('child_process').spawn;
let path = require('path');

const ARCH_MAPPING = {
    "ia32": "386",
    "x64": "amd64",
    "arm64": "arm64"
};

// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING = {
    "darwin": "darwin",
    "linux": "linux",
    "win32": "windows",
};

let exe = `../dist/mokapi-${PLATFORM_MAPPING[process.platform]}-${ARCH_MAPPING[process.arch]}/mokapi`
if (process.platform === 'win32') {
    exe += '.exe'
}
const exePath = path.resolve(__dirname, exe)
const mokapi = spawn(exePath, process.argv.slice(2), {cwd: process.cwd()})

mokapi.stdout.on('data', function(data) {
    console.log(data.toString())
})

mokapi.stderr.on('data', function(data) {
    console.log(data.toString())
})

