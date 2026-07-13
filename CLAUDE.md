# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository overview

`underwire-go` holds simple game backend services and utilities. It is two independent Go modules that share a data contract (an OpenAPI spec) but do not import each other:

- **`jij-service/`** — a Go HTTP backend (module `jij-service`) generated from `jij-service/api/openapi.yaml`, backed by MongoDB.
- **`oapi-codegen-client-ue/`** — a standalone Go CLI (module `oapi-codegen-client-ue`) that reads that same `openapi.yaml` and generates Unreal Engine (C++) client code from custom `text/template` templates.
- **`jij-compose/`** — Docker Compose files to run `jij-service` + MongoDB locally.

There is no root Go module; each subdirectory (`jij-service`, `oapi-codegen-client-ue`) is built and tested independently, with its own `go.mod`.

## jij-service

An HTTP API server built on `oapi-codegen`'s strict std-http-server generator.

### Architecture

Request flow: `main.go` wires everything together, then a layered pipeline handles each call:

```
net/http -> middleware.LoggingMiddleware -> api.HandlerFromMux -> api.StrictHandler -> api.Server (server.go) -> domain.ProfileService -> dal.ProfileDal (Mongo)
```

- `api/openapi.yaml` is the source of truth for the API surface (schemas + paths). `api/oapi-config.yaml` configures generation (models + std-http-server + strict-server, single-file output).
- `api/server.gen.go` is **generated** from the spec via the `go:generate` directive in `api/server.go` (`go tool oapi-codegen -config oapi-config.yaml openapi.yaml`) — do not hand-edit it; edit `openapi.yaml` and regenerate.
- `api/server.go` implements the `StrictServerInterface` methods generated into `server.gen.go` (`Server` struct holds a `ProfileService`). This is where request objects are translated to domain calls and domain results are wrapped into `...ResponseObject` types.
- `domain/profile_service.go` holds business logic (currently thin — e.g. `PatchProfileComponent` has a `// TODO Create if needed, else patch`).
- `dal/profile_dal.go` defines the `ProfileDal` interface (`Create`/`Retrieve`/`Update`/`Delete` over `ProfileComponentEntity`); `dal/profile_dal_mongo.go` is the only implementation, backed by a `profile-components` Mongo collection filtered by `nom` (component name) + `userID`.
- `configuration/service_config.go` loads `ServiceConfiguration` (core + Mongo settings) as JSON from the path in env var `JIJ_SERVICE_CONFIG_URI` — there is no default/fallback path, the server fails fast if it's unset.
- `middleware/QueryLogging.go` logs every request method/path/body and response duration; it buffers and restores the request body to keep it readable by downstream handlers.

### Commands (run from `jij-service/`)

```
go build .                                              # build the service
go run .                                                # run locally (requires JIJ_SERVICE_CONFIG_URI env var pointing at a JSON config file)
go generate ./api                                       # regenerate server.gen.go from openapi.yaml after editing the spec
go test ./...                                           # run tests
```

Local config JSON shape (see `configuration/service_config.go`):
```json
{ "core": { "environment": "..." }, "mongo": { "clientUri": "...", "database": "..." } }
```

Local stack via Docker Compose (`jij-compose/`, uses `docker-compose.yml` + `docker-compose.override.yml`): spins up `jij-service` (mapped to host port 5020) and a `mongodb` container, injecting `JIJ_SERVICE_CONFIG_URI` from the override file.

`jij-service-profile.http` has example requests for manual testing (note: some paths there, e.g. `/api/profile/...`, are stale/aspirational and don't match the current `openapi.yaml` paths like `/profile/{profileId}/component/{componentId}` — trust `openapi.yaml` over the `.http` file).

## oapi-codegen-client-ue

A code generator (not affiliated with the real `oapi-codegen` project beyond borrowing its name) that turns an OpenAPI 3 spec into Unreal Engine C++ source: header structs for each schema plus a `ServiceClient` class with one async method per path+verb.

### Architecture

`main.go` is the entry point: it parses `-spec=`, `-tmpl=`, `-out=` from `key=value` CLI args (custom parsing, not `flag`), loads the spec with `kin-openapi`, and drives generation in three phases:

1. `oapi.ExtractComponents` (in `oapi/oapi.go`) walks `components.schemas` and produces `OapiStructTypeInfo` (name + fields + external type dependencies) for each schema; `tmpl.GenerateStruct` renders each through `templates/struct.tmpl` into `<out>/<layer>/<SchemaName>.h`.
2. `oapi.ExtractServiceEndpoints` walks `paths`, and for each verb produces a `ServiceEndpoint` (PascalCase name derived from path+verb via `utils.ToPascalCase`, printf-style path with `%s` placeholders for path params, query/path params, request body type, response body type). `tmpl.GenerateServiceClient` renders `templates/serviceclient.h.tmpl` / `.cpp.tmpl` into `ServiceClient.h`/`.cpp`.
3. `tmpl.GenerateServiceClientHelpers` renders the layer-generic `serviceclienthelper.h/.cpp.tmpl` and `serviceclientmodels.h.tmpl` (response-info wrapper types, HTTP helper glue) into `ServiceClientHelper.h/.cpp` and `ServiceClientModels.h`.

Key type-mapping logic lives in `oapi.getUnrealTypeInfo`: OpenAPI `$ref` → `F<SchemaName>` struct pointer; `array` → `TArray<T>`; `object` with `additionalProperties` → `TMap<FString, T>`; plain `object` → `FJsonObjectWrapper`; `string`/`integer`/`number`/`boolean` → `FString`/`int32`/`float`/`bool`; `date-time` format → `FDateTime`. `IsEngineType` marks built-in Unreal types so `oapi.RemoveEngineTypes`/`Unique`/`Cleanup`/`Sort` can compute the minimal, deduplicated, sorted list of custom-struct forward declarations/includes each generated file needs.

`context.TemplateGenerationContext{TemplateDir, OutputDir, Layer}` is threaded through every generation call; `Layer` (currently always `"serviceclient"`) becomes both the output subdirectory and the Unreal module/include-path prefix baked into generated headers (e.g. `#include "{{.Layer}}/ServiceClientHelper.h"`).

Templates (`templates/*.tmpl`) are plain Go `text/template` files — when adding a new OpenAPI feature (e.g. a new schema shape or parameter style), the type-mapping change belongs in `oapi/oapi.go`, and the C++ text shape belongs in the relevant `.tmpl` file.

### Commands (run from `oapi-codegen-client-ue/`)

```
go build .
./oapi-codegen-client-ue.exe -spec="../jij-service/api/openapi.yaml" -tmpl=templates -out="./generated"
```

`generate.bat` wraps exactly this (build then run against the sibling `jij-service` spec, output to `./generated`). There is no `go generate`/Makefile — regeneration is always this manual two-step. The `generated/` directory is checked-in generator output, not hand-edited source; the `oapi-codegen-client-ue.exe` binary is also checked in (Windows-first workflow, run via `generate.bat`).

## Cross-cutting notes

- Both modules require newer Go toolchains (`go 1.26.x` in `go.mod`) — check your local Go version if builds fail on syntax.
- The two modules are only coupled through the shared `openapi.yaml`; changing that spec affects both `go generate` output in `jij-service` and the Unreal client generated by `oapi-codegen-client-ue` — regenerate both after editing it.
- License headers: source files carry an MIT copyright header (template in `.license`); match this style in new files.