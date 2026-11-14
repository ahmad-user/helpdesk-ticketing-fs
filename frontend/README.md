# Helpdesk Ticketing System - Frontend Application

[![Vue Version](https://img.shields.io/badge/Vue-3.5%2B-42b883?logo=vuedotjs)](https://vuejs.org)
[![Vite](https://img.shields.io/badge/Vite-5.4-646CFF?logo=vite)](https://vitejs.dev)
[![Pinia](https://img.shields.io/badge/Pinia-2.3-ffd859?logo=vuedotjs)](https://pinia.vuejs.org)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-3.4-38B2AC?logo=tailwindcss)](https://tailwindcss.com)
[![Axios](https://img.shields.io/badge/Axios-1.7-5A29E4?logo=axios)](https://axios-http.com)

A high-performance, enterprise-grade Single Page Application (SPA) built with **Vue 3 (Composition API & `<script setup>`)**, **Vite**, **Pinia**, **Vue Router**, and **Tailwind CSS**. Developed for the **Technical Test – Shift Engineer Fullstack at PT Simple Journey Indonesia**.

---

## 1. Features & Capabilities

- **Strict Role-Based Access Control (RBAC):** Dynamic interface adaptation for `requester`, `agent`, and `admin` roles with defense-in-depth routing guards.
- **Stateless JWT Authentication:** Automatic Bearer token injection, session re-hydration via `GET /api/me`, central 401 expiration handling with redirect loop safeguards.
- **Server-Driven Ticket Directory:** Real-time search with input debouncing, server-side pagination, sorting (`created_at`, `updated_at`, `priority`, `status`), and status/priority filters.
- **Lifecycle State Machine Matrix:** Status transitions strictly adhere to the technical specification matrix (`open` ➔ `in_progress` ➔ `resolved` ➔ `closed` / `reopen`) with interactive confirmation modals.
- **Triage & Specialist Assignment:** Self-assignment for Support Agents, candidate specialist selection for Administrators, and priority controls.
- **Discussions & Internal Notes Privacy:** Discussion threads with distinct styling for confidential support notes. Strict database-level isolation guarantees internal notes never leak to requesters.
- **Administrative User Management:** Paginated user directory, search, role modification (`requester`, `agent`, `admin`), and protection against demoting the last remaining administrator.
- **Production-Ready Docker Container:** Multi-stage build (`node:20-alpine` + `nginx:1.27-alpine`) serving static SPA assets with Nginx fallback routing on port `3000`.

---

## 2. Technology Stack

- **Framework:** Vue 3 (`vue@^3.5.13`) using `<script setup>` Composition API
- **Build Tool:** Vite (`vite@^5.4.14`, `@vitejs/plugin-vue@^5.2.1`)
- **State Management:** Pinia (`pinia@^2.3.1`)
- **Routing:** Vue Router (`vue-router@^4.5.0`)
- **HTTP Client:** Axios (`axios@^1.7.9`) with request/response interceptors
- **Styling:** Tailwind CSS (`tailwindcss@^3.4.17`, `autoprefixer@^10.4.20`, `postcss@^8.5.1`)
- **Icons:** Lucide Icons (`lucide-vue-next@^0.475.0`)

---

## 3. Project Architecture

```text
frontend/
├── public/
│   └── favicon.svg                               # Brand vector icon
├── src/
│   ├── assets/
│   │   └── main.css                              # Tailwind CSS directives & typography rules
│   ├── components/
│   │   ├── common/
│   │   │   ├── BaseBadge.vue                     # Reusable status/priority/role badge
│   │   │   ├── BaseButton.vue                    # Multi-variant button with loading state
│   │   │   ├── BaseCard.vue                      # Enterprise card container
│   │   │   ├── BaseInput.vue                     # Form text/password input with error feedback
│   │   │   ├── BaseSelect.vue                    # Normalized form dropdown
│   │   │   ├── BaseSpinner.vue                   # Loading spinner indicator
│   │   │   ├── BaseTextarea.vue                  # Multiline text input
│   │   │   ├── ConfirmModal.vue                  # Accessible confirmation dialog modal
│   │   │   ├── EmptyState.vue                    # Zero-data state placeholder
│   │   │   └── ErrorState.vue                    # Error boundary with retry action
│   │   └── layout/
│   │       ├── AppHeader.vue                     # Top header with user capsule & mobile trigger
│   │       └── AppSidebar.vue                    # Responsive sidebar navigation with role filtering
│   ├── composables/
│   │   └── useNotification.js                    # Reactive toast notification composable
│   ├── layouts/
│   │   ├── AppLayout.vue                         # Main dashboard layout wrapper
│   │   └── AuthLayout.vue                        # Centered layout for login/registration
│   ├── pages/
│   │   ├── auth/
│   │   │   ├── LoginPage.vue                     # Sign in with demo credential fill helpers
│   │   │   └── RegisterPage.vue                  # Public requester registration
│   │   ├── dashboard/
│   │   │   └── DashboardPage.vue                 # Live database metrics & recent tickets feed
│   │   ├── tickets/
│   │   │   ├── TicketCreatePage.vue              # Ticket submission form
│   │   │   ├── TicketDetailPage.vue              # Details, state machine actions & comments
│   │   │   └── TicketListPage.vue                # Filterable directory with server pagination
│   │   ├── users/
│   │   │   └── UserManagementPage.vue            # Administrator account & role management
│   │   └── NotFoundPage.vue                      # 404 fallback page
│   ├── router/
│   │   └── index.js                              # Route guards, auth checks & title sync
│   ├── services/
│   │   ├── api.js                                # Centralized Axios instance with interceptors
│   │   ├── authService.js                        # Authentication endpoint handlers
│   │   ├── commentService.js                     # Ticket comments & internal notes handlers
│   │   ├── ticketService.js                      # Ticket CRUD and transition handlers
│   │   └── userService.js                        # Administrative user directory handlers
│   ├── stores/
│   │   ├── auth.js                               # Pinia authentication & session store
│   │   └── ticket.js                             # Pinia ticket management & filter store
│   ├── utils/
│   │   └── formatters.js                         # Date, status, priority, and role formatters
│   ├── App.vue                                   # Dynamic layout resolver & toast container
│   └── main.js                                   # Application bootstrap
├── .dockerignore                                 # Container build ignores
├── .env                                          # Local environment variables
├── .env.example                                  # Environment template
├── Dockerfile                                    # Multi-stage production container
├── nginx.conf                                    # Nginx SPA fallback configuration (port 3000)
├── package.json                                  # NPM dependencies & scripts
├── tailwind.config.js                            # Tailwind theme configuration
└── vite.config.js                                # Vite bundler configuration
```

---

## 4. Prerequisites

- **Node.js:** Version 18 or higher (`node -v`)
- **NPM:** Version 9 or higher (`npm -v`)
- **Backend Service:** Running Go backend API at `http://localhost:8080/api`

---

## 5. How to Run Locally

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Configure environment:**
   Verify `.env` exists (copied from `.env.example`):
   ```env
   VITE_API_BASE_URL=http://localhost:8080/api
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```
   The application will be accessible at:
   - **Local URL:** `http://localhost:5173/`

4. **Compile production build:**
   ```bash
   npm run build
   ```

5. **Preview production build:**
   ```bash
   npm run preview
   ```

---

## 6. How to Run with Docker Compose

From the root project repository:
```bash
docker compose up --build
```
- **Frontend SPA:** `http://localhost:3000`
- **Backend API:** `http://localhost:8080`
- **Swagger Documentation:** `http://localhost:8080/docs`

---

## 7. Verified Demo Accounts

The database seeder provisions verified accounts with pre-hashed bcrypt passwords:

| Role | Email | Password | Permissions & Scope |
| :--- | :--- | :--- | :--- |
| **Admin** | `admin@example.com` | `Admin123!` | Full privileges, User Management, Role changes, All tickets, Internal notes |
| **Agent** | `agent@example.com` | `Agent123!` | View all tickets, Status transitions, Triage/assign, Internal notes |
| **Requester** | `requester@example.com` | `Requester123!` | Create tickets, View own tickets, Close own ticket, Reopen resolved |

---

## 8. Status Transition Matrix

| Current Status | Target Status | Permitted Roles | Description |
| :--- | :--- | :--- | :--- |
| **Open** | `in_progress` | Agent, Admin | Specialist claims issue and begins active investigation |
| **Open** | `closed` | Requester, Admin | Issue cancelled or self-resolved by requester |
| **In Progress** | `resolved` | Agent, Admin | Technical resolution implemented |
| **Resolved** | `closed` | Requester, Agent, Admin | Verified resolution accepted |
| **Resolved** | `open` | Requester | Issue recurred; returned to engineering queue |
| **Closed** | *(Terminal)* | *(None)* | Ticket closed permanently |
