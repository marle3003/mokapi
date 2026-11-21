---
title: Start Mokapi in CI and Wait for the Health Check
description: Start Mokapi in a CI pipeline as a background process or Docker container, wait until the health endpoint answers 200 before the tests run, and stop it afterwards.
---
# Start Mokapi in CI and Wait for the Health Check

Mokapi needs a moment to start. A fixed `sleep` either wastes time or is too short on a slow runner. Mokapi's
[health endpoint](/docs/configuration/health-check) lets you wait exactly as long as needed: poll it until it
answers `200`, then run your tests. If it never does, the job fails with a clear message.

## Wait for /health

By default, Mokapi serves `GET /health` on port `8080` and answers with `200` and `{"status":"healthy"}`. This
shell function polls the endpoint and fails after 30 seconds:

```bash
wait_for_mokapi() {
  url="${1:-http://localhost:8080/health}"
  for i in $(seq 1 30); do
    if curl -fs --max-time 2 "$url" > /dev/null; then
      echo "Mokapi is ready"
      return 0
    fi
    sleep 1
  done
  echo "Mokapi did not become healthy within 30 seconds" >&2
  return 1
}
```

`-f` makes curl return a non-zero exit code for HTTP errors, and `--max-time` keeps a single attempt from
hanging. If your curl is 7.71 or newer, one command does the same:

```bash
curl -fs --retry 30 --retry-delay 1 --retry-connrefused --retry-all-errors http://localhost:8080/health
```

`--retry-connrefused` retries while nothing listens on the port yet. `--retry-all-errors` also covers an empty reply,
which Docker's port forwarding can return in the first moments after the container starts. Without it, curl gives up
on that error immediately.

``` box=info title="Ready means listening"
Mokapi registers the health endpoint before it has read your specs, so `200` means that Mokapi's HTTP server
is up, not that every spec is loaded. For the first request of your tests, you can wait for a mocked endpoint
instead, for example `curl -fs http://localhost/api/v3/pet/1`.
```

## As a background process

Install Mokapi (see [Installation](/docs/get-started/installation.md)), start it in the background and wait for
it. Specs and scripts come from a directory in your repository, see
[Running Mokapi](/docs/configuration/running-mokapi.md).

```bash
mokapi ./mocks > mokapi.log 2>&1 &
MOKAPI_PID=$!

wait_for_mokapi || { cat mokapi.log; kill "$MOKAPI_PID"; exit 1; }

npm test
status=$?

kill "$MOKAPI_PID"
exit $status
```

Mokapi shuts down cleanly on `SIGTERM` and `SIGINT` and exits with code `0`. Printing the log when the wait fails
shows why Mokapi didn't start, for example an invalid flag or a port that is already in use. Saving the exit
code of your tests before the `kill` keeps the job result correct.

If Node.js is available on the runner, `npm install -g go-mokapi` installs the same binary as `mokapi`.

## As a Docker container

The official image `mokapi/mokapi` starts the same binary, see
[Docker and Docker Compose](/docs/configuration/use-cases/docker.md). Start it detached, wait, run the tests and
stop the container:

```bash
docker run -d --rm --name mokapi \
  -p 80:80 -p 8080:8080 \
  -v "$(pwd)/mocks:/mocks:ro" \
  mokapi/mokapi:latest /mocks

wait_for_mokapi || { docker logs mokapi; exit 1; }

npm test
status=$?

docker stop mokapi
exit $status
```

Pin a version tag such as `mokapi/mokapi:0.57.0` instead of `latest` if you want reproducible pipelines.
`docker stop` sends `SIGTERM`, and with `--rm` the container is removed afterwards.

## GitHub Actions

Steps of a job share the runner, so the container keeps running between steps. The wait step uses the curl
variant, so no helper script is needed:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Start Mokapi
        run: |
          docker run -d --rm --name mokapi \
            -p 80:80 -p 8080:8080 \
            -v "${{ github.workspace }}/mocks:/mocks:ro" \
            mokapi/mokapi:latest /mocks

      - name: Wait for Mokapi
        run: |
          curl -fs --retry 30 --retry-delay 1 --retry-connrefused --retry-all-errors \
            http://localhost:8080/health

      - name: Run tests
        run: npm test

      - name: Mokapi log
        if: failure()
        run: docker logs mokapi

      - name: Stop Mokapi
        if: always()
        run: docker stop mokapi
```

`if: always()` stops the container even when the tests fail. The tutorial
[Running Mokapi in a CI/CD Pipeline](/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline) walks through a complete
project with spec, script and tests.

## When port 8080 is taken

The dashboard, the REST API and the health endpoint share port `8080`. If the application under test also uses
`8080`, move Mokapi with [`--api-port`](/docs/configuration/static/cli-flags#api-port). `--health-port` has its own
default of `8080` and doesn't follow `--api-port`, so set it too, otherwise Mokapi starts a health listener on
`8080` again:

```bash
mokapi --api-port 9090 --health-port 9091 --health-path /healthz ./mocks
```

Because the ports differ, Mokapi starts a dedicated HTTP listener for the health endpoint. Poll
`http://localhost:9091/healthz` and, when you run in Docker, publish that port too. The same options exist as
environment variables (`MOKAPI_API_PORT`, `MOKAPI_HEALTH_PORT`, `MOKAPI_HEALTH_PATH`).
