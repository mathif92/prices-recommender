# Prices Recommender

Collects hotel prices from SerpAPI and stores them for analysis and recommendations.

## Features

- **Collection pipeline** — fetches hotel data from SerpAPI, maps responses to domain types, and persists to PostgreSQL
- **REST API** — browse hotels with prices, reviews, and ratings; manage settings; trigger collection
- **React frontend** — searchable hotel catalog with price ranges, detail views, and settings management
- **Integration tests** — full test suite against a disposable PostgreSQL via testcontainers

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- npm (for frontend development)
- SerpAPI API key

## Getting Started

### 1. Start PostgreSQL

```shell
docker compose up -d
```

### 2. Run database migrations

You'll need `psql` (PostgreSQL client):

```shell
for f in internal/dal/migration/*.sql; do
  sed '/^-- +goose Down$/,$d' "$f" | PGPASSWORD=prices psql -U prices -d prices_recommender -h localhost
done
```

Or use [goose](https://github.com/pressly/goose):

```shell
go install github.com/pressly/goose/v3/cmd/goose@latest
DATABASE_URL="postgres://prices:prices@localhost:5432/prices_recommender?sslmode=disable"
goose -dir internal/dal/migration postgres "$DATABASE_URL" up
```

### 3. Load sample data (optional)

```shell
PGPASSWORD=prices psql -U prices -d prices_recommender -h localhost -f internal/dal/data/sample_data.sql
```

### 4. Run the application

```shell
go run ./cmd/app
```

The server starts on `http://localhost:8080`. The frontend is served at `/` and the API at `/api/`.

### 5. Trigger data collection

```shell
curl -X POST http://localhost:8080/api/collect
```

## Frontend Development

For live reload during development:

```shell
cd frontend
npm install
npm run dev
```

This starts Vite's dev server on `http://localhost:5173`, proxying `/api` requests to the Go backend.

To build the frontend for production:

```shell
cd frontend && npm run build
```

The built assets are embedded into the Go binary via `//go:embed`.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/hotels` | List hotels (supports `?page=`, `?limit=`, `?location=`, `?with_prices=true`) |
| `GET` | `/api/hotels/{id}` | Hotel detail with ratings, reviews, and prices |
| `GET` | `/api/hotels/{id}/reviews` | Reviews for a hotel |
| `GET` | `/api/hotels/{id}/prices` | Price history |
| `GET` | `/api/hotels/{id}/ratings` | Rating history |
| `GET` | `/api/vacations` | List vacations (`?year=` to filter) |
| `POST` | `/api/collect` | Trigger data collection from SerpAPI |
| `GET` | `/api/settings` | List user settings |
| `POST` | `/api/settings` | Create/update a setting `{key, value}` |
| `GET` | `/api/settings/{key}` | Get a specific setting |
| `PUT` | `/api/settings/{key}` | Update a setting's value |
| `DELETE` | `/api/settings/{key}` | Delete a setting |

## Configuration

All configuration is via environment variables (see `.env`):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERP_API_KEY` | — | SerpAPI API key |
| `SERP_API_BASE_URL` | `https://serpapi.com/search` | SerpAPI base URL |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `prices` | PostgreSQL user |
| `DB_PASSWORD` | `prices` | PostgreSQL password |
| `DB_NAME` | `prices_recommender` | PostgreSQL database name |
| `HTTP_PORT` | `8080` | HTTP server port |

## Testing

```shell
# Run all integration tests (requires Docker)
go test -count=1 -timeout=300s ./...
```

Tests use testcontainers to spin up a disposable PostgreSQL per package. No external migration tool needed — migrations are embedded via `//go:embed`.

## Project Structure

```
cmd/app/                     # Application entry point
frontend/                    # React frontend (Vite)
  embed.go                   # Embeds built dist into Go binary
  src/
    App.jsx / App.css        # Root layout + navigation
    HotelsPage.jsx/.css      # Hotel search, list, detail
    SettingsPage.jsx/.css    # Settings management
    api.js                   # API client
internal/dal/
  migration/                 # Database migrations (SQL)
  data/sample_data.sql       # Sample data for local dev
  testhelpers/               # Testcontainers bootstrap + embedded migrations
pkg/
  api/                       # HTTP handlers (hotels, settings, collect, etc.)
  client/serpapi/            # SerpAPI HTTP client + response types
  collector/                 # Collector interface + composite collector
  collector/serpapi/         # SerpAPI collector (fetches + maps + saves)
  job/                       # Job runner (reads settings, iterates dates/locations)
  repositories/              # Database repository (CRUD + queries)
  types/                     # Domain types (HotelData, Hotel, etc.)
```

## Architecture

1. **SerpAPI Client** (`pkg/client/serpapi/`) — raw API calls returning typed responses
2. **Mapper** (`pkg/collector/serpapi/mapper.go`) — converts SerpAPI response types to domain types
3. **Collector** (`pkg/collector/`) — orchestrates multiple data sources, returns `[]types.HotelData`
4. **Repository** (`pkg/repositories/`) — persists data to PostgreSQL via sqlx with upsert semantics
5. **Job** (`pkg/job/`) — reads user settings, iterates over locations and date ranges, triggers collection
6. **API** (`pkg/api/`) — HTTP handlers for browsing data and triggering collection
7. **Frontend** (`frontend/`) — React SPA with modern CSS, served by the Go binary

## Sample Data

`hotels_riviera_maya_example.json` contains a sample SerpAPI response for reference during development.
