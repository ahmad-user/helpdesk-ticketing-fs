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

### Requirements

- Docker Engine
- Docker Compose

### 1. Clone the Repository

```bash
git clone https://github.com/ahmad-user/helpdesk-ticketing-fs.git
cd helpdesk-ticketing-fs
```

### 2. Configure Environment

Create the Docker environment file:

```bash
cp .env.example .env.docker
```

Edit `.env.docker` and configure the required environment variables, especially `JWT_SECRET`.

Use a strong, random secret for JWT authentication. Never commit environment files containing secrets to the repository.

### 3. Build and Start the Application

Run this command from the project root:

```bash
docker compose --env-file .env.docker up --build -d
```

Check the status of the services:

```bash
docker compose --env-file .env.docker ps
```

View application logs:

```bash
docker compose --env-file .env.docker logs -f
```

Database migrations and seed data are configured to run automatically when the backend starts.

## Application Access

| Service              | URL                          |
| -------------------- | ---------------------------- |
| Frontend             | http://localhost:3000        |
| Backend Health Check | http://localhost:8081/health |
| API Documentation    | http://localhost:8081/docs   |

The frontend uses Nginx to proxy API requests to the backend.

Open the frontend URL to access the application.

## Demo Accounts

| Role      | Email                 | Password      |
| --------- | --------------------- | ------------- |
| Admin     | admin@example.com     | Admin123!     |
| Agent     | agent@example.com     | Agent123!     |
| Requester | requester@example.com | Requester123! |

These accounts are intended for development and testing purposes.

## Documentation

- [Backend Documentation](backend/README.md)
- [Frontend Documentation](frontend/README.md)
- [API Documentation](http://localhost:8081/docs)

## Project Structure

```text
.
├── backend/
├── frontend/
├── docker-compose.yml
├── .env.example
└── README.md
```

## Stop the Application

Stop and remove the application containers:

```bash
docker compose --env-file .env.docker down
```

To remove the database volume and all stored database data:

```bash
docker compose --env-file .env.docker down -v
```

**Warning:** The `down -v` command permanently deletes the database data stored in the Docker volume.
