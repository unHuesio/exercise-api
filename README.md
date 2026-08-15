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

## CI

The project uses GitHub Actions to run linting first, then tests, and finally a build check.
