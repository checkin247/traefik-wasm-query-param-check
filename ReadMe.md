
[![CI (WASM plugin)](https://github.com/checkin247/traefik-wasm-query-param-check/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_ORG/traefik-wasm-query-param-check/actions/workflows/ci.yml)


# Query Parameter validation for traefik
> uses wasm


Validates a query parameter against an allow-list.

Proceed if query parameter is set and its value is in the list of allowed values.

## Configuration

### Plugin configuration

```yaml
experimental:
  plugins:
    queryToken:
      moduleName: github.com/checkin247/traefik-wasm-query-param-check
      version: v1.0.2
```

## Middleware configuration
```yaml
http:
  middlewares:
    token-check:
      plugin:
        qptoken:
          paramName: "Token"
          allowedValues: ["your-prod-token"]
          denyStatus: 401
```


## build 

### Create or update go.sum

```bash
# On PowerShell use ${PWD.Path} (or use the included helper script below)
docker run --rm -v "${PWD.Path}":/work -w /work/src golang:1.22-bookworm /usr/local/go/bin/go mod tidy
```

### build with docker

```bash
# Prefer the cross-platform helper script which will use a local tinygo when
# available or fall back to a Docker-based tinygo. The script will also try to
# run `wasm-opt` (Binaryen) if present to optimize the generated wasm. You can
# override the optimizer path with the WASMOPT environment variable.

# PowerShell
./scripts/build-wasm.ps1

# POSIX (bash)
./scripts/build-wasm.sh
```
this will also move it to plugins directory used by docker-compose

## Test locally

```bash
docker-compose up -d
```
then
```bash
curl -i 'http://localhost:80/?Token=my-secret'
curl -i 'http://localhost:80/?Token=not-mysecret'
```

## Tests

Unit tests live next to the code in `src/` and are run by default in CI.

Run unit tests locally:

```powershell
Push-Location src
go test ./...
Pop-Location
```

Integration tests (end-to-end) are build-tagged and run separately so they don't
execute during normal unit test runs or CI. To run the local integration tests:

```powershell
# Run integration-tagged tests in src (requires a running Traefik instance)
Push-Location src
go test -tags=integration -run TestLocalIntegration -v ./...
Pop-Location
```

Use the `LOCAL_TEST_BASE` environment variable to point the integration test at a
different base URL (default: `http://localhost:80/`).

## Debugging / Dev mode

When troubleshooting the plugin in a cluster or local Traefik instance you can
enable a small set of runtime debug logs and decision points by setting
`devMode: true` in the plugin config. This enables extra logging which helps
observe why requests are allowed or denied.

Example dynamic configuration (Traefik):

```yaml
http:
  middlewares:
    token-check:
      plugin:
        qptoken:
          paramName: "Token"
          allowedValues: ["my-secret"]
          denyStatus: 401
          devMode: false
```

Notes:
- When the plugin is built with TinyGo and run in the http-wasm host, logs are
  forwarded to the host via `handler.Host.Log(...)` and will appear in Traefik
  or the host runtime logs depending on your deployment.
- Log levels use the host's `api.LogLevel` values; the plugin emits debug/info
  messages for decision points when `devMode` is enabled.
- Keep `devMode` off in production unless you need diagnostics; logs may contain
  request data that you should avoid exposing in production environments.



## License

Free to use under GPL-3.0, see LICENSE

## Remarks
Mostly auto generated
