# Sarfya – Run fya'ot a sar

I developed this tool to be able to efficiently find the usages of Na'vi words in canon and approved examples.

This is the library half of it with the core logic, the rest will be in [sarfya-service](https://github.com/gissleh/sarfya-service).

## License

The GPL license is no longer in effect since the GPL dependency is in the other package.

The text used in `data/` is the property of their original authors, 
and are attributed under the `source` property of the YAML documents.

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

The data is not included here, but you can build it with the other repository or download it from https://sarfya.vmaple.dev/data.json

### Adapters

#### `placeholderdictionary`

This is run alongside the main dictionary to handle placeholders like `X-ìl`.
It's also used for names that aren't in the dictionary.

#### `jsonstorage`

An indexed storage backend for `service` that can be loaded and saved as a JSON.
This is meant for the production server to be bundled with the whole compiled dataset.



