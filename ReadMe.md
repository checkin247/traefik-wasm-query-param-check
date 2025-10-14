
[![CI (WASM plugin)](https://github.com/checkin247/traefik-wasm-query-param-check/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_ORG/traefik-wasm-query-param-check/actions/workflows/ci.yml)


# Query Parameter validation for traefik
> uses wasm


Validates a query parameter against an allow-list.

Proceed if query parameter is set and its value is in the list of allowed values.

## Config (dynamic)
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
docker run --rm -v "$PWD":/work -w /work/src golang:1.22-bookworm /usr/local/go/bin/go mod tidy
```

### build with docker

```bash
docker run --rm -v "$PWD":/work -w /work/src tinygo/tinygo:0.34.0 tinygo build -o /work/plugin.wasm -scheduler=none --no-debug -target=wasi .
```

### move to build dir

```bash
MODULE_PATH=github.com/checkin247/traefik-wasm-query-param-check
mkdir -p plugins-local/src/$MODULE_PATH
cp plugin.wasm .traefik.yml plugins-local/src/$MODULE_PATH/
```

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


## License

Free to use under GPL-3.0, see LICENSE

## Remarks
Mostly auto generated
