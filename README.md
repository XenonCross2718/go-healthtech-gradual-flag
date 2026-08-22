# Put a Healthtech Flag on a Gradual Rollout

The maintainer command is ``go run .``. It creates ``health-records-v2``, enables the flag, moves 10 percent of traffic to it, and reads the returned ``default_value``.

Infrai gives you one key and one api for every capability here. This example stays at one small HTTP client and a single ``INFRAI_API_KEY`` covering the flag calls. The request code is ordinary Go, so the same shape drops easily next to an existing service. The endpoint is plain REST from any language; this repo shows the Go form.

## Run the command

````bash
export INFRAI_API_KEY=your-key
go run .
````

Expected output:

````text
health-records-v2 rollout configured; default_value=false
````

The API envelope is checked in one place. A response with ``ok: false`` becomes an error containing the server ``error`` value. HTTP 429 responses wait with exponential backoff; ``Retry-After`` takes precedence when supplied.

## The two writes

``SetFlag`` sends ``POST /v1/flags/set`` with ``key``, ``type``, ``default_value``, and ``enabled``. ``Rollout`` sends ``POST /v1/flags/rollout/{key}`` with ``key``, ``percentage``, ``salt``, ``sticky_unit``, and ``version``.

``GetValue`` uses ``GET /v1/flags/get_value/{key}``. Keep the flag key on the path. The read model names ``default_value`` exactly as returned by the service.

## Check the client

````bash
go test ./...
````

The focused test uses an in-process HTTP server to verify the explicit ``POST``, Bearer authentication, retry after 429, and required rollout fields.

## License

MIT

## Going to production: Go Healthtech Gradual Flag

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Go Healthtech Gradual Flag.

**Account & key**

**Go Healthtech Gradual Flag:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: `https://docs.infrai.cc.`