# Query Parameter validation for traefik
> uses wasm


Validates a query parameter against an allow-list.

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
MODULE_PATH=github.com/checkin247/traefik-wasm-token-check
mkdir -p build/plugins-local/src/$MODULE_PATH
cp plugin.wasm .traefik.yml build/plugins-local/src/$MODULE_PATH/
```



## Mis


[![CI (WASM plugin)](https://github.com/checkin247/traefik-wasm-token-check/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR_ORG/traefik-wasm-token-check/actions/workflows/ci.yml)

