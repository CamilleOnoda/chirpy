# Chirpy — Twitter-like API in Go

A simple Twitter-like backend API that lets users create accounts, post short messages ("chirps"), and interact with a database through HTTP endpoints.

Built in Go with a PostgreSQL backend.

---

## What it does

Chirpy exposes a REST-style API where users can:

- Register and authenticate
- Create, retrieve, and delete chirps
- Manage authentication tokens (login, refresh, revoke)
- Handle external events via webhooks

The API is designed to mirror real backend systems: it uses proper HTTP semantics, persistent storage, and token-based authentication.

---

## Prerequisites

**Go 1.21+**
Download from https://go.dev/dl or install via your package manager.

**PostgreSQL**

macOS:
```bash
brew install postgresql@16
brew services start postgresql@16
```

Ubuntu/Debian:
```bash
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

Windows: use the official installer.

---

## Installation

**1. Clone the repository**
```bash
git clone https://github.com/CamilleOnoda/chirpy
cd chirpy
```

**2. Set up the database**
```bash
createdb chirpy
```

**3. Configure environment variables**

Create a `.env` file:

```
B_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your-secret-key
```
**4. Run migrations**
```bash
goose up
```

**5. Start the server**
```bash
go run main.go
```

The server will start locally (e.g. `http://localhost:8080`).

---

## Quick start

Using a REST client (VSCode RESTClient, curl, Postman):

```http
# Create a user
POST /api/users

# Login
POST /api/login

# Create a chirp
POST /api/chirps

# Get all chirps
GET /api/chirps

# Delete a chirp
DELETE /api/chirps/{id}

# Refresh token
POST /api/refresh

# Revoke token
POST /api/revoke
```

You can use the provided `RESTClient.http` file to test endpoints directly.

---

## Endpoints overview

### Users & Auth

| Method | Endpoint       | Description          |
|--------|----------------|----------------------|
| POST   | /api/users     | Create user          |
| PUT    | /api/users     | Update user          |
| POST   | /api/login     | Authenticate         |
| POST   | /api/refresh   | Refresh access token |
| POST   | /api/revoke    | Revoke refresh token |

### Chirps

| Method | Endpoint           | Description      |
|--------|--------------------|------------------|
| POST   | /api/chirps        | Create chirp     |
| GET    | /api/chirps        | Get all chirps   |
| GET    | /api/chirps/{id}   | Get chirp by ID  |
| DELETE | /api/chirps/{id}   | Delete chirp     |

### Webhooks

| Method | Endpoint              | Description           |
|--------|-----------------------|-----------------------|
| POST   | /api/polka/webhooks   | Handle external events|

---

## Project structure
```
chirpy/
├── main.go                    # Server setup and routing
├── handler_*.go               # HTTP handlers (one per endpoint)
├── internal/
│   ├── auth/                  # Authentication logic (tokens, hashing)
│   ├── database/              # Generated DB layer (sqlc)
│   └── static/                # Static files (HTML, assets)
├── sql/
│   ├── queries/               # Raw SQL queries
│   └── schema/                # Database migrations
├── metrics.go                 # Middleware (request tracking)
├── readiness.go               # Health check endpoint
├── reset.go                   # Dev-only reset logic
├── RESTClient.http            # API testing file
├── sqlc.yaml                  # sqlc configuration
├── go.mod / go.sum            # Dependencies
└── README.md

```

---

## Tech stack

- **Go** — HTTP server (`net/http`), JSON handling
- **PostgreSQL** — persistent storage
- **sqlc** — type-safe SQL → Go code generation
- **goose** — database migrations
- **Argon2** — password hashing
- **Token-based authentication** (access + refresh tokens)
- **lib/pq** — PostgreSQL driver

---

## Notes

- Handlers are split per endpoint to keep the API surface explicit and easy to navigate
- SQL is written manually and compiled with `sqlc` for clarity and type safety
- Authentication is isolated from handlers to keep concerns separated
- Webhooks follow an idempotent design and safely ignore irrelevant events (`204 No Content`)

---

## What I learned

- How HTTP APIs are structured without frameworks
- Why correct status codes and headers matter
- How token-based authentication works in practice
- How to design database-backed APIs
- How real systems handle external events (webhooks)

---

## Next improvements

- Pagination for chirps
- Rate limiting
- Better test coverage
- Structured logging
- Deployment (AWS)