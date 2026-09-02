# Exercise API

[![CI](https://github.com/unHuesio/exercise-api/actions/workflows/ci.yml/badge.svg)](https://github.com/unHuesio/exercise-api/actions/workflows/ci.yml)

A Go REST API for managing gym exercises and routines, built with Gin, MongoDB, and Casbin RBAC.

## Features

- JWT-based authentication for user-facing clients
- Role-based authorization with Casbin
- Exercise and routine CRUD endpoints
- MongoDB-backed persistence
- CI pipeline for linting, testing, and build validation

## Local development

```bash
go mod download
go run .
```

Set the required environment variables before running the app:

- `MONGO_URI`
- `JWT_SECRET`

The app will also load a local `.env` file automatically when present.

## API documentation

With the application running, the interactive ReDoc reference is available at
[`/docs`](http://localhost:8080/docs). The source-controlled OpenAPI 3.0 document
is served at [`/openapi.yaml`](http://localhost:8080/openapi.yaml) and lives in
[`docs/openapi.yaml`](docs/openapi.yaml).

Public registration and login accept a Google OpenID Connect `id_token`. A successful
`/login` response contains the application JWT; send it as
`Authorization: Bearer <token>` for protected operations. Casbin policies determine
which authenticated subjects may use each protected endpoint.

When adding or changing a route, update `docs/openapi.yaml` in the same change so the
served contract and ReDoc reference remain accurate.

## CI

The project uses GitHub Actions to run linting first, then tests, and finally a build check.
