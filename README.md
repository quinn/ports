# ports

Small CLI for stable local service port allocation.

`ports` stores mappings in `~/.ports.json` keyed by namespace and service:

- If a mapping already exists for `namespace/service`, it returns the existing port.
- Otherwise it finds an available port, saves it, and prints it.

## Install

```bash
go build -o ports .
```

Place `ports` on your `PATH` (or invoke it as `./ports`).

## Usage

```bash
ports <service>
```

- `<service>` is required (for example: `api`, `web`, `postgres`).
- Namespace defaults to the current directory name.

Examples:

```bash
# Uses current directory as namespace
ports api
```

## Example `~/.ports.json`

```json
{
  "my-app": {
    "api": 3000,
    "web": 3001
  }
}
```

## `justfile` Example

```just

API_PORT := `ports api`

export API_URL := "http://localhost:" + API_PORT # env API_URL is read by web

run-api:
  PORT={{ API_PORT }} go run ./cmd/api

run-web:
  PORT=$(ports web) go run ./cmd/web
```

This lets `just run-api` and `just run-web` reuse stable ports from the same namespace.
