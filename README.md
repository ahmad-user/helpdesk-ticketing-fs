# Helpdesk Ticketing System

A fullstack Helpdesk Ticketing System built with Go, Vue 3, and PostgreSQL.

## Features

- JWT Authentication
- Role-Based Access Control (Requester, Agent, Admin)
- Ticket Management and Assignment
- Ticket Priority and Status Workflow
- Comments and Internal Notes
- Search, Filtering, and Pagination
- Admin User Management

## Tech Stack

- **Backend:** Go, Gin
- **Frontend:** Vue 3, Vite, Pinia
- **Database:** PostgreSQL
- **Authentication:** JWT, bcrypt
- **Deployment:** Docker, Docker Compose, Nginx

## Getting Started

**Requirements:** Docker Engine and Docker Compose.

1. Clone the repository:

```bash
git clone https://github.com/ahmad-user/helpdesk-ticketing-fs.git
cd helpdesk-ticketing-fs
```

2. Configure environment:

```bash
cp .env.example .env
```

Configure the required environment variables, especially `JWT_SECRET`.

3. Start the application:

```bash
docker compose up --build
```

## Application Access

| Service           | URL                        |
| ----------------- | -------------------------- |
| Frontend          | http://localhost:3000      |
| Backend           | http://localhost:8080      |
| API Documentation | http://localhost:8080/docs |

## Demo Accounts

| Role      | Email                 | Password      |
| --------- | --------------------- | ------------- |
| Admin     | admin@example.com     | Admin123!     |
| Agent     | agent@example.com     | Agent123!     |
| Requester | requester@example.com | Requester123! |

## Documentation

- [Backend Documentation](backend/README.md)
- [Frontend Documentation](frontend/README.md)
- [API Documentation](http://localhost:8080/docs)

## Project Structure

```text
.
├── backend/
├── frontend/
├── docker-compose.yml
├── .env.example
└── README.md
```

## Stop Application

```bash
docker compose down
```

To remove database volumes and stored data:

```bash
docker compose down -v
```
