# Sarfya – Run fya'ot a sar

I developed this tool to be able to efficiently find the usages of Na'vi words in canon and approved examples.

The svelte frontend is held together by duct tape, and it is only meant for editing the content in the `data/` folder for now.

## License

The code under `ports/fwewdictionary/` and any package depending on these (only `cmd/...` within the project) is licensed under GPL2 (LICENSE-GPL.txt).

The text used in `data/` is the property of their original authors, and are attributed under the `source` property.

The remaining code and the annotations in `data/` falls under the ISC license (LICENSE-ISC.txt).

## Running & Building

### Dev (read/write) server

You need Go >= 1.18 to run the backend.

```bash
go run ./cmd/sarfya-dev-server/
```

You can then load the read-only frontend at http://localhost:8080

If you intend to use the editor, you need the svelte frontend for now.
Run these commands to get started with it. I have not tested with Node.JS < v18
Follow the instructions on the screen to open the page.

```bash
cd frontend/
npm install
env VITE_ENABLE_EDITOR=true VITE_BACKEND_URL="http://localhost:8080" npm run dev
```

### Prod (readonly) server

The minimum you need to run the application in a readonly form is this.
You can then 

```bash
go build -ldflags "-w -s" ./cmd/sarfya-prod-server/
go run ./cmd/sarfya-generate-json/
zip -r sarfya-prod-server.zip sarfya-prod-server data-compiled.json ~/.fwew/dictionary-v2.txt
```

### Lambda

You can set this up with a `API Gateway proxy event`.
It is currently not working with most of the operator shorthands when filtering,
but it will work.

```bash
go run ./cmd/sarfya-generate-json/
go build -o bootstrap -ldflags "-w -s" ./cmd/sarfya-aws-lambda/
zip -r sarfya-aws-lambda.zip bootstrap data-compiled.json dictionary-v2.txt
```

## Docker

Build the Dockerfile, it uses the `cmd/sarfya-prod-server` as the main command and bundles fwew with it.

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

#### `templfrontend`

A WIP templ-based frontend that's faster to host. 
It's a bit coupled with the `webapi` package, however.
