# Helpdesk Ticketing System – Frontend

A web-based Helpdesk Ticketing System built with Vue 3, designed to manage support tickets, user roles, and ticket workflows.

Developed for **Technical Test – Shift Engineer Fullstack at PT Simple Journey Indonesia**.

## Features

- Role-Based Access Control (Requester, Agent, Admin).
- JWT authentication and session management.
- Ticket creation and management.
- Search, filtering, sorting, and pagination.
- Ticket assignment and priority management.
- Ticket comments and private internal notes.
- Ticket status workflow and transition validation.
- Admin user management and role updates.
- Responsive interface for desktop and mobile.

## Tech Stack

- **Framework:** Vue 3 (Composition API)
- **Build Tool:** Vite
- **State Management:** Pinia
- **Routing:** Vue Router
- **HTTP Client:** Axios
- **Styling:** Tailwind CSS
- **Icons:** Lucide Vue
- **Deployment:** Docker & Nginx

## Getting Started

### Prerequisites

- Node.js 18+
- npm 9+
- Backend API running

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

Copy `.env.example` to `.env` and configure:

```env
VITE_API_BASE_URL=http://localhost:8080/api
```

### 3. Run Development Server

```bash
npm run dev
```

Frontend will be available at:

`http://localhost:5173`

### 4. Production Build

```bash
npm run build
```

Preview the production build:

```bash
npm run preview
```

## Run with Docker Compose

From the root project directory:

```bash
docker compose up --build
```

Application URLs:

- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8080`
- API Documentation: `http://localhost:8080/docs`

## Demo Accounts

| Role      | Email                 | Password      |
| --------- | --------------------- | ------------- |
| Admin     | admin@example.com     | Admin123!     |
| Agent     | agent@example.com     | Agent123!     |
| Requester | requester@example.com | Requester123! |

## User Roles

- **Requester:** Create tickets, view own tickets, comment, and close or reopen eligible tickets.
- **Agent:** View and manage tickets, assign staff, update priority and status, and write internal notes.
- **Admin:** Manage all tickets, users, and roles.
