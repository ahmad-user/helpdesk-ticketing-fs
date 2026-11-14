# Helpdesk Ticketing System - Enterprise Backend API

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?logo=docker)](https://www.docker.com)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-85EA2D?logo=openapi-initiative)](https://swagger.io)

An enterprise-grade, high-performance Backend Helpdesk Ticketing System developed in Go (Golang) following Clean Layered Architecture, strict Role-Based Access Control (RBAC), and stateless JWT Bearer token authentication.

Developed for **Technical Test – Shift Engineer Fullstack at PT Simple Journey Indonesia**.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Technology Stack](#2-technology-stack)
3. [Architecture & Design Principles](#3-architecture--design-principles)
4. [Folder Structure](#4-folder-structure)
5. [Prerequisites](#5-prerequisites)
6. [Environment Configuration](#6-environment-configuration)
7. [Database Setup & Schema](#7-database-setup--schema)
8. [Database Migration Instructions](#8-database-migration-instructions)
9. [Database Seeding Instructions](#9-database-seeding-instructions)
10. [How to Run Locally](#10-how-to-run-locally)
11. [How to Run with Docker Compose](#11-how-to-run-with-docker-compose)
12. [Demo Accounts](#12-demo-accounts)
13. [API Endpoint Directory](#13-api-endpoint-directory)
14. [Swagger / OpenAPI Documentation](#14-swagger--openapi-documentation)
15. [Testing & Quality Assurance](#15-testing--quality-assurance)
16. [Security Architecture & Considerations](#16-security-architecture--considerations)

---

## 1. Project Overview

The Helpdesk Ticketing System provides a secure REST API platform for managing IT support tickets across an organization. It features distinct user roles (`requester`, `agent`, `admin`), ticket lifecycle state machines, assignment workflows, rich discussion comments, confidential internal notes for support personnel, and full administrative user management.

### Key Capabilities:
- **Authentication & RBAC:** Secure public registration (forced requester), bcrypt password hashing, signed JWT issuance, and dynamic database-backed role synchronization.
- **Ticket Management:** Complete CRUD, fine-grained state transition matrix, priority management, agent assignment, and advanced search/filter/sorting.
- **Comments & Internal Notes:** Real-time discussion threads with strict database-level filtering preventing confidential internal notes from ever leaking to requesters.
- **Administrative Control:** User listing with pagination, search, role modification, and protection preventing the accidental demotion of the last remaining system administrator.

---

## 2. Technology Stack

- **Language:** Go (v1.22+ / v1.27.3)
- **HTTP Web Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin) (`v1.12.0`)
- **Database:** PostgreSQL 16
- **Database Driver:** `github.com/lib/pq`
- **Database Migrations:** `github.com/golang-migrate/migrate/v4`
- **Security & Cryptography:** `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5`
- **Configuration Management:** `github.com/joho/godotenv`
- **Validation:** `github.com/go-playground/validator/v10`
- **Identifiers:** `github.com/google/uuid` (UUIDv4)
- **Containerization:** Docker (Multi-stage build) & Docker Compose (v2)
- **API Documentation:** OpenAPI 3.0 / Swagger UI (embedded viewer)

---

## 3. Architecture & Design Principles

The application strictly implements **Layered Clean Architecture** and **Inversion of Control (Dependency Injection)**. Each layer holds specific, decoupled responsibilities:

```
┌────────────────────────────────────────────────────────┐
│                   HTTP Transport Layer                 │
│         Gin Router, Middlewares (Auth, RBAC, CORS)     │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                      Handler Layer                     │
│    Request Binding, Param Parsing, Response Envelopes   │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                      Service Layer                     │
│   Business Logic, State Machine, Security Validations  │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                    Repository Layer                    │
│    PostgreSQL Queries, Transactions, Joins (No N+1)    │
└───────────────────────────┬────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────┐
│                     Database Layer                     │
│         PostgreSQL Database Engine & Connection Pool    │
└────────────────────────────────────────────────────────┘
```

### Architectural Guarantees:
- **Context Propagation:** All database and service operations take `context.Context` for cancellation and deadline propagation.
- **Zero N+1 Queries:** Relational entities (`Requester`, `Assignee`, `Author`) are eagerly joined using SQL `JOIN` clauses.
- **Standardized Response Envelope:** Every API endpoint returns a uniform JSON schema (`success`, `message`, `data`, `pagination`, `errors`).

---

## 4. Folder Structure

```text
backend/
├── cmd/
│   ├── api/
│   │   └── main.go                  # Application entrypoint & Dependency Injection wiring
│   ├── migrate/
│   │   └── main.go                  # Database migrations CLI (up & down)
│   └── seed/
│       └── main.go                  # Database seeder CLI
├── docs/
│   └── openapi.yaml                 # Complete OpenAPI 3.0 specification
├── internal/
│   ├── config/                      # Environment configuration loader
│   ├── database/                    # PostgreSQL connection pool & migration runner
│   ├── handler/                     # HTTP transport handlers (Auth, Ticket, Comment, User, Health)
│   ├── middleware/                  # HTTP middlewares (Logger, Recovery, Auth, RBAC, CORS)
│   ├── model/                       # Domain models, request payloads & response DTOs
│   ├── repository/                  # Database repository interfaces & SQL implementations
│   ├── routes/                      # Route registration and endpoint groups
│   ├── service/                     # Business domain logic and security rules
│   └── validator/                   # Input validation error formatting
├── migrations/                      # Reversible database migrations (.up.sql & .down.sql)
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_tickets_table.up.sql
│   ├── 000002_create_tickets_table.down.sql
│   ├── 000003_create_comments_table.up.sql
│   └── 000003_create_comments_table.down.sql
├── seeds/
│   └── seed.go                      # Idempotent database seeder logic
├── tests/                           # Unit & integration test suites
│   ├── auth_test.go
│   ├── health_test.go
│   ├── prompt5_test.go              # Complete 16-test suite for Prompt 5
│   ├── rbac_test.go
│   └── ticket_test.go
├── .env.example                     # Environment variables template
├── .gitignore                       # Git ignore specifications
├── Dockerfile                       # Multi-stage production container definition
├── docker-compose.yml               # Local multi-service compose configuration
├── entrypoint.sh                    # Container startup script (migrations, seed, api)
├── go.mod                           # Go module manifest
├── go.sum                           # Go module checksums
└── README.md                        # Documentation
```

---

## 5. Prerequisites

Before installing, ensure the following tools are installed on your machine:
- **Go:** Version 1.22 or higher (`go version`)
- **PostgreSQL:** Version 16 or higher (running locally or accessible via network)
- **Docker & Docker Compose:** (Optional, for containerized execution)

---

## 6. Environment Configuration

Copy the sample environment template:

```bash
cp .env.example .env
```

### Environment Variable Reference:

| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `APP_NAME` | `helpdesk-ticketing-system` | Application name identifier |
| `APP_ENV` | `development` | Environment mode (`development`, `production`, `test`) |
| `APP_PORT` | `8080` | Port for the HTTP server to listen on |
| `APP_HOST` | `0.0.0.0` | Bind IP host address |
| `DB_HOST` | `localhost` (or `postgres` in Docker) | PostgreSQL host address |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `helpdesk_db` | PostgreSQL database name |
| `DB_SSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |
| `DB_MAX_OPEN_CONNS` | `25` | Maximum active database connections in pool |
| `DB_MAX_IDLE_CONNS` | `10` | Maximum idle connections in pool |
| `DB_CONN_MAX_LIFETIME` | `15m` | Maximum connection lifetime |
| `JWT_SECRET` | *(required in prod)* | Cryptographic HMAC-SHA256 secret key |
| `JWT_EXPIRATION_HOURS` | `24` | Token lifespan in hours |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:5173` | Comma-separated allowed frontend origins |
| `AUTO_MIGRATE` | `true` | Automatically run migrations up on container start |
| `AUTO_SEED` | `true` | Automatically seed demo accounts on container start |

---

## 7. Database Setup & Schema

1. Connect to PostgreSQL and create the database:
   ```sql
   CREATE DATABASE helpdesk_db;
   ```

2. The schema consists of 3 relational tables:
   - **`users`**: UUID primary key, unique email, bcrypt hash, role check (`requester`, `agent`, `admin`), timestamps, and role/email indexes.
   - **`tickets`**: UUID primary key, foreign key references to `users(id)` for `requester_id` (CASCADE) and `assignee_id` (SET NULL), status and priority check constraints, and composite indexes.
   - **`comments`**: UUID primary key, foreign keys to `tickets(id)` and `users(id)` with CASCADE, `is_internal` boolean flag, and indexed `ticket_id`, `author_id`, `created_at`.

---

## 8. Database Migration Instructions

Database migrations are managed using SQL files with full reversible `up` and `down` scripts.

### Run Migrations Up (Apply all pending schemas):
```bash
go run ./cmd/migrate -direction=up
```

### Run Migrations Down (Roll back schema):
```bash
go run ./cmd/migrate -direction=down
```

---

## 9. Database Seeding Instructions

The seeder inserts bootstrap accounts for each user role idempotently (`ON CONFLICT (email) DO NOTHING`):

```bash
go run ./cmd/seed
```

---

## 10. How to Run Locally

1. **Verify dependencies:**
   ```bash
   go mod download
   ```

2. **Execute database migrations:**
   ```bash
   go run ./cmd/migrate -direction=up
   ```

3. **Seed demo users:**
   ```bash
   go run ./cmd/seed
   ```

4. **Start the API server:**
   ```bash
   go run ./cmd/api
   ```

   The server will start listening at `http://localhost:8080`.
   - Health check: `http://localhost:8080/health`
   - Interactive Swagger UI: `http://localhost:8080/docs`
   - Raw OpenAPI YAML: `http://localhost:8080/docs/openapi.yaml`

---

## 11. How to Run with Docker Compose

Running the entire backend and PostgreSQL database requires just **one command**:

```bash
docker compose up --build
```

### What happens automatically:
1. PostgreSQL container starts and establishes volume persistence (`postgres_data`).
2. PostgreSQL healthcheck verifies readiness (`pg_isready`).
3. Backend container compiles via multi-stage build and starts.
4. `entrypoint.sh` executes database migrations (`AUTO_MIGRATE=true`).
5. `entrypoint.sh` executes idempotent database seed (`AUTO_SEED=true`).
6. The Helpdesk API server begins serving traffic on port `8080`.

To stop the services:
```bash
docker compose down
```

To stop and remove database volumes (clean slate):
```bash
docker compose down -v
```

---

## 12. Demo Accounts

The database seeder provisions verified demo credentials with bcrypt-hashed passwords:

| Role | Name | Email | Password | Allowed Capabilities |
| :--- | :--- | :--- | :--- | :--- |
| **Admin** | System Administrator | `admin@example.com` | `Admin123!` | Full system privileges, User management, Role updates, All tickets, Internal notes |
| **Agent** | Default Support Agent | `agent@example.com` | `Agent123!` | View all tickets, Update status, Assign tickets, Change priority, Internal notes |
| **Requester** | Default Requester | `requester@example.com` | `Requester123!` | Create tickets, View own tickets, Close own tickets, Reopen resolved tickets, Public comments |

---

## 13. API Endpoint Directory

All endpoints under `/api` (except `/api/auth/register` and `/api/auth/login`) require the `Authorization: Bearer <token>` header.

### Health Check:
| Method | Path | Auth | Roles | Description |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/health` | Public | All | Service status, uptime, database connection state |

### Authentication & Profile:
| Method | Path | Auth | Roles | Description |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/auth/register` | Public | All | Register a new user (role forced to `requester`) |
| `POST` | `/api/auth/login` | Public | All | Login with email and password to receive JWT |
| `GET` | `/api/me` | Bearer | All | Retrieve current authenticated user profile |

### Tickets:
| Method | Path | Auth | Roles | Description |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/tickets` | Bearer | All | List tickets (Requesters see only their own; Agents/Admins see all). Supports pagination, status/priority/assignee filters, search, sorting |
| `POST` | `/api/tickets` | Bearer | All | Create a ticket (Requester ID forced from JWT context) |
| `GET` | `/api/tickets/:id` | Bearer | All | Get ticket details (IDOR protected for Requesters) |
| `PATCH` | `/api/tickets/:id` | Bearer | All | Update ticket status, priority, or assignee (State machine enforced) |
| `DELETE` | `/api/tickets/:id` | Bearer | None | Returns `405 Method Not Allowed` |

### Comments & Internal Notes:
| Method | Path | Auth | Roles | Description |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/tickets/:id/comments` | Bearer | All | List comments chronologically. Internal notes are filtered in SQL for Requesters |
| `POST` | `/api/tickets/:id/comments` | Bearer | All | Add comment. `is_internal: true` allowed only for Agent and Admin |

### Admin User Management:
| Method | Path | Auth | Roles | Description |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/users` | Bearer | Admin | List all registered users (paginated, search by name/email, password hash omitted) |
| `PATCH` | `/api/users/:id/role` | Bearer | Admin | Change user role (`requester`, `agent`, `admin`). Protects last admin |

---

## 14. Swagger / OpenAPI Documentation

Interactive Swagger documentation is served directly by the application:
- **Swagger UI:** [http://localhost:8080/docs](http://localhost:8080/docs)
- **OpenAPI Specification File:** [http://localhost:8080/docs/openapi.yaml](http://localhost:8080/docs/openapi.yaml) or located at `backend/docs/openapi.yaml`.

---

## 15. Testing & Quality Assurance

The codebase features comprehensive unit and integration tests covering positive flows, negative flows, boundary values, RBAC enforcement, and data privacy.

### Run All Tests:
```bash
go test -v -count=1 ./...
```

### Run Race Condition Detector:
```bash
go test -race -count=1 ./...
```

### Code Formatting & Static Analysis:
```bash
gofmt -s -w .
go vet ./...
go build ./...
```

### Test Coverage Summary:
- **`tests/auth_test.go`**: Registration validation, duplicate email conflict, password hashing verification, login credentials check, token generation and profile retrieval.
- **`tests/ticket_test.go`**: Ticket creation, IDOR prevention, assignee assignment rules, status transition matrix validation, pagination, filtering, search, sorting.
- **`tests/rbac_test.go`**: Missing token rejection, invalid signature rejection, expired token rejection, role inheritance verification.
- **`tests/prompt5_test.go`**: Complete 16-test suite for comments, internal notes privacy, user listing, role manipulation prevention, fail-closed invalidation of old tokens.

---

## 16. Security Architecture & Considerations

1. **Authentication & Token Management:**
   - Passwords hashed using bcrypt with automatic salt generation (`bcrypt.DefaultCost`).
   - JWT tokens signed using HMAC-SHA256 (`HS256`) with configurable secret and expiration.
   - Dynamic database role checking in `AuthMiddleware`: when a user's role is updated, existing JWT tokens immediately reflect the new permissions or are rejected with `403 Forbidden`.
   - **Fail-Closed Principle:** If database connection fails during authenticated token verification, requests are rejected immediately rather than failing open.

2. **Authorization & RBAC:**
   - Role hierarchy strictly enforced: `Admin` inherits `Agent` permissions. `Requester` is strictly isolated.
   - Requesters can only view, update, and comment on tickets where `ticket.requester_id == caller.id`.

3. **Data Privacy & Leak Prevention:**
   - `password_hash` is explicitly excluded from JSON serialization (`json:"-"`) and never sent in API responses.
   - Internal notes (`is_internal = TRUE`) are filtered at the database query level (`AND ($2 = TRUE OR c.is_internal = FALSE)`) preventing data leakage.

4. **Input Validation & SQL Injection Prevention:**
   - All database queries use parameterized placeholders (`$1, $2, ...`). Zero raw string interpolation.
   - Input fields validated for length, required status, and strict enumerations.

5. **Operational Safeguards:**
   - Protected against demoting the last remaining system administrator (`cannot demote the last remaining administrator`).
   - CORS middleware configured with a whitelist of permitted frontend origins.
   - Multi-stage Docker container runs under an unprivileged non-root user (`appuser`, UID 10001).
   - Sensitive credentials excluded from Git via `.gitignore`.
