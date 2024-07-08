# Sarfya – Run fya'ot a sar

I developed this tool to be able to efficiently find the usages of Na'vi words in canon and approved examples.

It is not yet ready for prime time, but I am sharing it for local testing and demonstration.

The web frontend is held together by duct tape, tsatseng lu hì'ang apxay!

## License

The project code and annotations in `./data` falls under the ISC license.
The text used in `./data` is the property of their original authors.

## Quick Start

You need Go >= 1.18 and Node.JS >= v1.18.

### 1. Start backend
```bash
go run ./cmd/sarfya-dev-server/
```

### 2. Start frontend

```bash
cd frontend/
npm install
env VITE_ENABLE_EDITOR=true VITE_BACKEND_URL="http://localhost:8080" npm run dev
```

### 3. Preview
Open a browser and navigate to http://localhost:8080

## Project structure

All 'business logic' and the main functionality is in the root of the project, and has zero external dependencies.
The goal of that is to facilitate integration into existing `fwew` services.

For a minimal use of the library, see `./cmd/fyasar-example`

### Adapters

The adapters are where the external dependencies come into play.

#### `fwewdictionary`

This hooks the `fyasar.Dictionary` interface up with `fwew`.

#### `placeholderdictionary`

This is run alongside `fwew` to handle placeholders like `X-ìl`.
It's also used for names that aren't in the dictionary.

#### `jsonstorage`

An indexed storage backend for `service` that can be loaded and saved as a JSON.
The plan is bundling a JSON with the binary and throw it up on a FaaS behind a CDN cache in the future.

#### `sourcestorage`

The storage backend for `service` ran locally. 
It modifies the relevant files in `./data`.

#### `webapi`

An (aspirationally) REST API for the frontend.