---
title: "Using Mokapi for Local Testing (Jest, Manual Testing, and Debugging)"
description: Learn to use Mokapi for local testing with Jest, manual testing, and debugging. Simulate mocked APIs without external dependencies.
icon: bi-play-circle
---

# Using Mokapi in Local Tests

Mokapi can be used for more than CI/CD — it's also perfect for local testing, debugging, and development workflows. 
Here’s how you can use it:

## Option 1: Run Mokapi as a Background Service

Start Mokapi manually and keep it running while you work. You can easily install Mokapi using a package manager for your system.

### Install Mokapi

- On macOS: Use Homebrew to install Mokapi:
```shell
brew tap marle3003/tap 
brew install mokapi
```

- On Windows: Use Chocolatey to install Mokapi:
```shell
choco install mokapi
```

- On Linux: Follow the installation instructions [here](/docs/guides/get-started/installation.md).

Once Mokapi is installed, you can start it from the command line:

### 💻 Run from Executable
```shell
mokapi api.yaml mock.ts
```

### 🐳 Use Docker
```shell
docker run --rm -p 80:80 -p 8080:8080 -v "$(pwd):/app" mokapi/mokapi:latest /app/user-api.yaml /app/mock.ts
```

Then your app, test suite, or browser can interact with the mocked APIs at http://localhost:80 and the dashboard
is accessible at http://localhost:8080

Good for:
- Running integration tests (Jest, Playwright, etc.)
- Manual testing with frontend or backend-to-backend apps
- Debugging workflows without needing a real backend
- Simulating error cases or slow responses

## Option 2: Start Mokapi Automatically in Test Code

Useful if you want everything bundled into a single test command.

```javascript tab=jest.setup.js
const { spawn } = require('child_process');
let mokapi;

beforeAll((done) => {
  mokapi = spawn('mokapi', ['user-api.yaml', 'mock.ts'], {
    stdio: 'inherit',
  });
  setTimeout(done, 2000);
});

afterAll(() => mokapi?.kill());
```

```javascript tab=test.js
describe('Mokapi service', () => {
  it('should respond with mock data', async () => {
    const res = await fetch('http://localhost:8080/users');
    expect(res.status).toBe(200);
  });
});
```

```javascript tab=jest.config.js
module.exports = {
    setupFilesAfterEnv: ['./jest.setup.js'],
};
```

## When Mokapi Might Not Be Needed

While Mokapi is great for integration and end-to-end testing, it’s generally not necessary for unit tests. Unit tests 
should be small, fast, and focused on testing the logic of your code, not external systems.

Instead of using Mokapi, you can mock external dependencies directly in your test code (e.g., with Jest mocks) to 
simulate responses from external services like APIs. This approach keeps your tests fast and deterministic, without 
relying on external services or introducing delays.

Use Mokapi for:
- Running integration or end-to-end tests where external interactions are involved.
- Simulating complex systems or error scenarios.

For unit tests, mocking dependencies directly is usually more effective for speed and isolation.

