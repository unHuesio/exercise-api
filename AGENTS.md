# AGENTS.md

This repository is a Go REST API for a gym/exercise application. It is built with Gin, MongoDB, and Casbin RBAC.

## Project purpose

The service exposes CRUD endpoints for exercises, routines, API keys, applications, permissions, and user authentication. It authenticates requests either with a JWT bearer token or an API key and authorizes actions through Casbin policies.

## High-level architecture

- main.go: application bootstrap, router setup, middleware registration, and protected/public route groups.
- config/config.go: environment loading for MongoDB and JWT secret.
- db/db.go: MongoDB connection lifecycle.
- handlers/: HTTP handlers for each domain.
- middleware/: Gin middleware for auth, API-key validation, JWT validation, RBAC authorization, and security headers.
- models/: persistence DTOs for database documents.
- domain/: domain types and repository interfaces.
- infrastructure/: concrete implementations, especially Mongo-backed persistence.
- config/rbac_model.conf: Casbin policy model.

## Runtime and local setup

- Language: Go.
- Entry point: main.go
- Build/run: `go run .`
- Required environment variables:
  - `MONGO_URI`
  - `JWT_SECRET`
- The app loads `.env` automatically via godotenv if present.

## Bootstrap flow

1. `config.Load()` reads `.env` and required environment variables.
2. `db.Connect(cfg.MongoURI)` opens the MongoDB client.
3. A Casbin enforcer is created using `config/rbac_model.conf` and the Mongo adapter.
4. Handlers are instantiated with the Mongo client and the enforcer.
5. Gin router is configured with global security and rate limiting middleware.
6. Public routes are mounted for registration/login and app token generation.
7. Protected routes are mounted under a single group with:
   - `Auth(...)`
   - `InferObjectAction()`
   - `Authorize(...)`

## Auth and authorization model

- Public endpoints: `/register`, `/login`, `/applications/token`, `/ping`
- Protected routes: all other endpoints, guarded by JWT or API-key middleware.
- `middleware.Auth` accepts either Authorization header JWT or `x-api-key`.
- JWT middleware populates request context with user info.
- API key middleware validates API keys against MongoDB collections.
- `middleware.InferObjectAction` sets the inferred `inferred_object` and `inferred_action` context values used by Casbin.
- `middleware.Authorize` enforces `sub, obj, act` permissions using the RBAC model.

## MongoDB conventions

- The app uses the Mongo database named `gym-app`.
- Collections are named in lower-case snake case in many places, e.g. `users`, `applications`, `api_keys`, `exercises`, `routines`.
- Entity structs use Mongo BSON tags and JSON tags. Example: `_id`, `email`, `api_key`, etc.
- Most handlers perform direct database operations inside their handler methods rather than through a repository layer.

## Important project conventions

- Keep handler logic close to the database access pattern already used in the repo.
- Prefer the same Gin response patterns:
  - `c.JSON(http.StatusOK, ...)`
  - `c.JSON(http.StatusCreated, ...)`
  - `c.JSON(http.StatusBadRequest, gin.H{"error": ...})`
  - `c.JSON(http.StatusUnauthorized, gin.H{"error": ...})`
  - `c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})`
- When changing routes or auth behavior, update both the router setup in `main.go` and the relevant middleware expectations.
- Casbin policy and role behavior is central to authorization; changes to roles or resource names should be validated against `config/rbac_model.conf` and call sites in `middleware/authorizeMiddleware.go`.

## Key files to read first

- [main.go](main.go)
- [config/config.go](config/config.go)
- [middleware/authMiddleware.go](middleware/authMiddleware.go)
- [middleware/authorizeMiddleware.go](middleware/authorizeMiddleware.go)
- [handlers/authentication.go](handlers/authentication.go)
- [handlers/exercises.go](handlers/exercises.go)
- [config/rbac_model.conf](config/rbac_model.conf)

## Working guidelines for agent tasks

- Follow the repo’s existing Gin + MongoDB pattern rather than introducing unrelated frameworks.
- Keep the architecture layered in the same shape as the existing code: router -> middleware -> handler -> Mongo collection.
- Preserve environment-vs-code configuration and the Casbin authorization flow when modifying security behavior.
- If a task is not obvious, start by checking the route in `main.go`, then the matching handler in `handlers/`, and finally the middleware that enforces access control.
- Prefer minimal changes that match the repository’s style and naming conventions.
