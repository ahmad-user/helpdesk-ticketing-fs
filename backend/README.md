# Helpdesk Ticketing System – Backend

REST API for an internal Helpdesk Ticketing System, built with Go and PostgreSQL.

Developed for **Technical Test – Shift Engineer Fullstack at PT Simple Journey Indonesia**.

## Features

- JWT authentication and password hashing (bcrypt).
- Role-Based Access Control (Requester, Agent, Admin).
- Ticket management, assignment, priority, and status transitions.
- Ticket comments and private internal notes.
- Search, filtering, sorting, and pagination.
- Admin user management and role updates.
- Database migrations and demo data seeding.

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Gin
- **Database:** PostgreSQL 16
- **Authentication:** JWT
- **Password Hashing:** bcrypt
- **API Documentation:** OpenAPI 3.0 / Swagger
- **Containerization:** Docker & Docker Compose

## Project Structure

```text
backend/
├── cmd/
├── docs/
├── internal/
├── migrations/
├── seeds/
├── tests/
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md
```

## Getting Started

### Environment Configuration

```bash
cp .env.example .env
```

Configure database credentials and `JWT_SECRET` in the `.env` file.

### Run with Docker Compose

```bash
docker compose up --build
```

Stop services:

```bash
docker compose down
```

### Run Locally

Make sure Go and PostgreSQL are installed and configured.

```bash
go mod download
go run ./cmd/migrate -direction=up
go run ./cmd/seed
go run ./cmd/api
```

Backend runs at `http://localhost:8080`.

## Demo Accounts

| Role      | Email                 | Password      |
| --------- | --------------------- | ------------- |
| Admin     | admin@example.com     | Admin123!     |
| Agent     | agent@example.com     | Agent123!     |
| Requester | requester@example.com | Requester123! |

## API Documentation

- Swagger UI: `http://localhost:8080/docs`
- OpenAPI specification: `docs/openapi.yaml`

## Testing

Run all tests:

```bash
go test -v -count=1 ./...
```

Additional checks:

```bash
go vet ./...
go build ./...
```

## User Roles

- **Requester:** Create tickets, view own tickets, comment, and close or reopen eligible tickets.
- **Agent:** Manage tickets, assign staff, update priority and status, and write internal notes.
- **Admin:** Full ticket management and user role administration.
