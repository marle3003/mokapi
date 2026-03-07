#!/usr/bin/env node
const { spawn } = require('child_process');
const fs = require('fs');

// Mapping Node platform + arch to scoped platform packages
function resolveBinary() {
    const platform = process.platform;
    const arch = process.arch;

    let pkgName;

    switch (platform) {
        case 'linux':
            pkgName = arch === 'arm64'
                ? '@go-mokapi/linux-arm64'
                : '@go-mokapi/linux-x64';
            break;
        case 'darwin':
            pkgName = arch === 'arm64'
                ? '@go-mokapi/darwin-arm64'
                : '@go-mokapi/darwin-x64';
            break;
        case 'win32':
            pkgName = arch === 'arm64'
                ? '@go-mokapi/win32-arm64'
                : '@go-mokapi/win32-x64';
            break;
        default:
            console.error(`Unsupported platform/arch: ${platform}/${arch}`);
            process.exit(1);
    }

    try {
        // Each platform package exports the full path to the binary
        const binaryPath = require(pkgName);
        if (!fs.existsSync(binaryPath)) {
            console.error(`Binary not found at ${binaryPath}`);
            process.exit(1);
        }
        return binaryPath;
    } catch (err) {
        console.error(`Platform package ${pkgName} is not installed.`);
        console.error(`Run 'npm install' to install the correct platform binary.`);
        process.exit(1);
    }
}

// Get the binary path for current platform
const exe = resolveBinary();

// Spawn Mokapi process with arguments from CLI
const mokapi = spawn(exe, process.argv.slice(2), { cwd: process.cwd() });

// Forward stdout/stderr
mokapi.stdout.on('data', data => process.stdout.write(data));
mokapi.stderr.on('data', data => process.stderr.write(data));

// Exit with the same code as Mokapi
mokapi.on('close', code => process.exit(code));