# Helpdesk Ticketing System (Monorepo)

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)
[![Vue 3](https://img.shields.io/badge/Vue-3.5%2B-4FC08D?logo=vue.js)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?logo=docker)](https://www.docker.com)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-85EA2D?logo=openapi-initiative)](https://swagger.io)

An enterprise-grade, high-performance Helpdesk Ticketing System built as a monorepo featuring a Go (Golang) REST API backend and a Vue 3 (Composition API `<script setup>`) frontend. Orchestrated via Docker Compose with PostgreSQL 16.

Developed for **Technical Test – Shift Engineer Fullstack at PT Simple Journey Indonesia**.

---

## Daftar Isi / Table of Contents

1. [Deskripsi Aplikasi](#1-deskripsi-aplikasi)
2. [Tech Stack](#2-tech-stack)
3. [Arsitektur Project](#3-arsitektur-project)
4. [Prasyarat (Prerequisites)](#4-prasyarat-prerequisites)
5. [Cara Menjalankan Menggunakan Docker Compose](#5-cara-menjalankan-menggunakan-docker-compose)
6. [Cara Menjalankan Manual (Local Development)](#6-cara-menjalankan-manual-local-development)
7. [Environment Variables](#7-environment-variables)
8. [Akun Seed Demo](#8-akun-seed-demo)
9. [Aturan Bisnis & Transisi Status Tiket](#9-aturan-bisnis--transisi-status-tiket)
10. [Daftar Endpoint API](#10-daftar-endpoint-api)
11. [Cara Menjalankan Database Migration](#11-cara-menjalankan-database-migration)
12. [Cara Menjalankan Test](#12-cara-menjalankan-test)
13. [Asumsi dan Batasan Implementasi](#13-asumsi-dan-batasan-implementasi)

---

## 1. Deskripsi Aplikasi

Helpdesk Ticketing System adalah platform manajemen tiket dukungan pelanggan dan operasional IT internal. Aplikasi menyediakan alur kerja terstruktur untuk pelaporan masalah teknis, penugasan staf (*assignment*), penyesuaian prioritas, diskusi interaktif (*comments*), catatan rahasia internal tim staf (*internal notes*), serta kontrol akses pengguna berbasis peran (*Role-Based Access Control* / RBAC).

### Fitur Utama:
- **Autentikasi & Otorisasi:** Registrasi publik (selalu menghasilkan role `requester`), login berbasis JWT stateless, verifikasi masa kedaluwarsa & signature, hash password menggunakan bcrypt, dan sinkronisasi role langsung ke database.
- **Role-Based Access Control (RBAC):** Tiga peran utama: `requester`, `agent`, dan `admin`.
- **Manajemen Tiket (CRUD):** Pembuatan tiket, daftar tiket dengan pencarian, filter status & prioritas, sorting aman (kolom ter-whitelist), dan paginasi.
- **State Machine Status Tiket:** Validasi transisi status tiket yang ketat di backend sesuai matriks bisnis.
- **Diskusi & Catatan Internal (Internal Notes):** Requester dan staf dapat bertukar pesan publik; staf (agent & admin) dapat menambahkan catatan internal rahasia yang difilter langsung pada level query database sehingga tidak pernah bocor ke requester.
- **Manajemen Pengguna (Admin):** Tambah pengguna, edit data, ganti role pengguna, hapus pengguna, serta perlindungan transaksi ACID agar tidak dapat menurunkan atau menghapus administrator terakhir (*last admin demotion protection*).
- **Frontend SPA Responsif:** Menggunakan Vue 3 Composition API `<script setup>`, Pinia store, Tailwind CSS, Lucide icons, navigasi role-aware, loading/empty/error state di setiap halaman, dan adaptif untuk desktop maupun mobile.

---

## 2. Tech Stack

### Backend:
- **Bahasa & Runtime:** Go 1.22+ (mendukung Go 1.27.x)
- **Web Framework:** Gin Web Framework (`github.com/gin-gonic/gin`)
- **Database:** PostgreSQL 16
- **Database Driver:** `github.com/lib/pq`
- **Database Migrations:** `github.com/golang-migrate/migrate/v4`
- **Security & JWT:** `golang.org/x/crypto/bcrypt`, `github.com/golang-jwt/jwt/v5`
- **Validation:** `github.com/go-playground/validator/v10`
- **UUID:** `github.com/google/uuid` (UUIDv4)
- **API Spec:** OpenAPI 3.0 (Swagger UI terintegrasi di `/docs`)

### Frontend:
- **Framework:** Vue 3 (Composition API dengan `<script setup>`)
- **Build Tool:** Vite 5
- **State Management:** Pinia 2
- **Routing:** Vue Router 4 (dengan navigation guards & role checks)
- **Styling:** Tailwind CSS 3 & PostCSS
- **HTTP Client:** Axios (otomatis inject JWT Bearer & interceptor 401)
- **Icons:** Lucide Vue Next

### Orchestration & Deployment:
- **Containerization:** Docker & Docker Compose v2
- **Web Server / Reverse Proxy:** Nginx 1.27 Alpine

---

## 3. Arsitektur Project

Struktur monorepo diorganisasikan secara modular:

```text
.
├── backend/                         # Go REST API backend
│   ├── cmd/
│   │   ├── api/main.go              # Server entrypoint & Dependency Injection
│   │   ├── migrate/main.go          # CLI runner database migration
│   │   └── seed/main.go             # CLI runner database seeder
│   ├── docs/openapi.yaml            # Spesifikasi OpenAPI 3.0
│   ├── internal/
│   │   ├── config/                  # Pengaturan environment & validasi JWT secret
│   │   ├── database/                # Koneksi PostgreSQL & migration engine
│   │   ├── handler/                 # HTTP handlers (Auth, Ticket, Comment, User, Health)
│   │   ├── middleware/              # Middlewares (Auth, RBAC, CORS, Recovery, Logger)
│   │   ├── model/                   # Domain entities, request payloads & DTOs
│   │   ├── repository/              # SQL queries & database transactions
│   │   ├── routes/                  # Definisi rute HTTP & grouping Gin
│   │   ├── service/                 # Aturan bisnis, state machine & validasi RBAC
│   │   └── validator/               # Formatter pesan error validasi input
│   ├── migrations/                  # Script SQL migration up/down yang reversibel
│   ├── seeds/seed.go                # Logika seeder user awal (idempoten)
│   ├── tests/                       # Unit dan integration tests backend
│   ├── Dockerfile                   # Multi-stage build image backend
│   ├── entrypoint.sh                # Container startup: auto-migrate, auto-seed, exec api
│   └── go.mod / go.sum
├── frontend/                        # Vue 3 SPA frontend
│   ├── src/
│   │   ├── components/              # Reusable UI components (Card, Button, Input, Modal, dll)
│   │   ├── composables/             # Shared composables (notifikasi, dll)
│   │   ├── layouts/                 # AppLayout & AuthLayout
│   │   ├── pages/                   # Halaman Auth, Dashboard, Tickets, User Management
│   │   ├── router/                  # Definisi rute & navigation guards
│   │   ├── services/                # Axios API service client
│   │   ├── stores/                  # Pinia stores (auth, ticket)
│   │   └── utils/                   # Formatters dan konstanta
│   ├── Dockerfile                   # Multi-stage build image frontend dengan Nginx
│   ├── nginx.conf                   # Konfigurasi Nginx SPA fallback & proxy /api/
│   └── package.json
├── docker-compose.yml               # Orkestrasi Docker untuk PostgreSQL, Backend, Frontend
├── .env.example                     # Contoh template environment variables
├── .env                             # Environment variables lokal
└── README.md                        # Dokumentasi menyeluruh
```

---

## 4. Prasyarat (Prerequisites)

Untuk menjalankan via Docker Compose:
- **Docker Engine:** Versi 20.10 atau lebih baru
- **Docker Compose:** Versi 2.0 atau lebih baru

Untuk menjalankan manual di host:
- **Go:** Versi 1.22 atau lebih baru (`go version`)
- **Node.js & npm:** Node.js v18+ dan npm v9+ (`node -v && npm -v`)
- **PostgreSQL:** Versi 16 atau lebih baru berjalan di port 5432

---

## 5. Cara Menjalankan Menggunakan Docker Compose

Satu perintah untuk membangun dan menjalankan seluruh service:

```bash
docker compose up --build
```

Perintah ini akan secara otomatis:
1. Menjalankan container PostgreSQL 16 dan menunggu hingga *healthy* via `pg_isready`.
2. Membangun binary Go backend, menjalankan migrasi database (`up`), melakukan seeding user awal, dan memulai server API di port `8080`.
3. Membangun bundle frontend Vue 3 dan menyajikannya via Nginx di port `3000`.

### Mengakses Aplikasi:
- **Frontend Web UI:** [http://localhost:3000](http://localhost:3000)
- **Backend API:** [http://localhost:8080/api](http://localhost:8080/api)
- **Health Check:** [http://localhost:8080/health](http://localhost:8080/health)
- **Swagger UI / API Docs:** [http://localhost:8080/docs](http://localhost:8080/docs)

Untuk menghentikan service:
```bash
docker compose down
```

Untuk menghapus volume data database (reset data):
```bash
docker compose down -v
```

---

## 6. Cara Menjalankan Manual (Local Development)

### Langkah 1: Siapkan Database PostgreSQL
Pastikan server PostgreSQL aktif dan buat database:
```sql
CREATE DATABASE helpdesk_db;
```

### Langkah 2: Jalankan Migrasi dan Seed Backend
Masuk ke direktori `backend/`:
```bash
cd backend

# Pastikan file .env tersedia (atau salin dari .env.example)
cp .env.example .env

# Jalankan database migrations up
go run ./cmd/migrate -direction=up -path=./migrations

# Jalankan database seeder
go run ./cmd/seed

# Jalankan server backend
go run ./cmd/api
```
Backend akan aktif di `http://localhost:8080`.

### Langkah 3: Jalankan Frontend
Buka terminal baru dan masuk ke direktori `frontend/`:
```bash
cd frontend

# Pastikan file .env tersedia
cp .env.example .env

# Install dependensi npm
npm install

# Jalankan server development Vite
npm run dev
```
Frontend development server akan aktif di `http://localhost:5173` (atau port yang ditampilkan Vite) dan terhubung ke backend di `http://localhost:8080/api`.

---

## 7. Environment Variables

Template environment variable disediakan di `.env.example`.

| Variabel | Default Docker | Keterangan |
| :--- | :--- | :--- |
| `APP_NAME` | `helpdesk-ticketing-system` | Nama aplikasi |
| `APP_ENV` | `development` | Mode aplikasi (`development`, `production`, `test`) |
| `APP_PORT` | `8080` | Port listen backend HTTP server |
| `APP_HOST` | `0.0.0.0` | Bind IP host address |
| `DB_HOST` | `postgres` (atau `localhost`) | Host PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | Username database |
| `DB_PASSWORD` | `postgres` | Password database |
| `DB_NAME` | `helpdesk_db` | Nama database |
| `DB_SSLMODE` | `disable` | Mode SSL PostgreSQL (`disable`, `require`) |
| `JWT_SECRET` | *(secret key dev)* | Secret key penandatanganan JWT HMAC-SHA256 (min. 32 char di production) |
| `JWT_EXPIRATION_HOURS` | `24` | Masa berlaku token dalam jam |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:5173` | Daftar origin frontend yang diizinkan |
| `AUTO_MIGRATE` | `true` | Otomatis menjalankan migrasi database saat container start |
| `AUTO_SEED` | `true` | Otomatis melakukan seeding akun demo saat container start |
| `FRONTEND_PORT` | `3000` | Port host untuk frontend |
| `VITE_API_BASE_URL` | `http://localhost:8080/api` | Base URL API yang diakses browser sisi client |

> [!NOTE]
> Pada mode `APP_ENV=production`, backend mewajibkan `JWT_SECRET` dikonfigurasi secara eksplisit dengan panjang minimal 32 karakter dan tidak boleh menggunakan secret default dev. Backend akan menolak start jika secret tidak aman.

---

## 8. Akun Seed Demo

Database seeder secara idempoten menyediakan tiga akun demo siap pakai:

| Role | Email | Password | Hak Akses |
| :--- | :--- | :--- | :--- |
| **Admin** | `admin@example.com` | `Admin123!` | Akses penuh seluruh tiket, assignment, ubah priority, transisi status, kelola user & role |
| **Agent** | `agent@example.com` | `Agent123!` | Lihat semua tiket, self-assignment / ubah assign, ubah priority, catatan internal, transisi status |
| **Requester** | `requester@example.com` | `Requester123!` | Buat tiket baru, lihat & komentar hanya pada tiket miliknya, tutup/buka kembali tiket miliknya |

*(Di halaman login web, tersedia tombol klik-cepat untuk mengisi akun demo di atas secara otomatis).*

---

## 9. Aturan Bisnis & Transisi Status Tiket

Backend menegakkan matriks state machine status tiket berikut:

| Dari Status | Ke Status | Role yang Diizinkan |
| :--- | :--- | :--- |
| **Open** | **In Progress** | `agent`, `admin` |
| **In Progress** | **Resolved** | `agent`, `admin` |
| **Resolved** | **Closed** | `requester`, `agent`, `admin` |
| **Resolved** | **Open** | `requester` |
| **Open** | **Closed** | `requester`, `admin` |

### Aturan Bisnis Penting:
1. **Transisi di luar tabel di atas ditolak** oleh backend dengan HTTP 400 Bad Request.
2. **Transisi yang tidak diizinkan untuk role terkait ditolak** dengan HTTP 403 Forbidden.
3. **Status `closed` adalah status terminal**: tidak dapat berpindah ke status lain kecuali tiket yang `resolved` dibuka kembali ke `open` oleh `requester`.
4. **Requester hanya dapat melihat dan mengomentari tiket miliknya sendiri**.
5. **Agent dan Admin dapat melihat seluruh tiket**.
6. **Catatan internal (`is_internal = true`)**: hanya dapat dibuat dan dilihat oleh `agent` dan `admin`. Difilter langsung pada level SQL query database.
7. **Assignment & Priority**: Hanya `agent` dan `admin` yang dapat mengubah prioritas tiket dan melakukan penugasan staf. Requester ditolak dengan HTTP 403.
8. **Validasi Assignee**: Pengguna yang ditugaskan harus merupakan user aktif dengan role `agent` atau `admin`.
9. **Perlindungan Admin Terakhir (*Last Admin Protection*)**: Backend menggunakan PostgreSQL advisory lock dan row lock (`FOR UPDATE`) dalam database transaction untuk mencegah demosi atau penghapusan administrator terakhir, termasuk saat terjadi *race condition* (request bersamaan).

---

## 10. Daftar Endpoint API

Semua respons endpoint terbungkus dalam format standar:
```json
{
  "success": true,
  "message": "Deskripsi hasil",
  "data": {},
  "pagination": { "total": 10, "page": 1, "limit": 10, "total_pages": 1 }
}
```

### Public & Health:
- `GET /health` – Status kesehatan aplikasi dan koneksi PostgreSQL
- `GET /docs` – Swagger UI API Documentation
- `GET /docs/openapi.yaml` – OpenAPI 3.0 YAML specification

### Autentikasi:
- `POST /api/auth/register` – Registrasi publik (role otomatis dipaksa `requester`)
- `POST /api/auth/login` – Login user & pengembalian token JWT Bearer
- `GET /api/me` – Profil user terautentikasi (memerlukan Header `Authorization: Bearer <token>`)

### Manajemen Tiket (Memerlukan JWT):
- `POST /api/tickets` – Buat tiket baru (default: status `open`, requester terikat ke pemanggil)
- `GET /api/tickets` – Daftar tiket (Mendukung query: `page`, `limit`, `status`, `priority`, `search`, `sort_by`, `order`, `assignee_id`)
- `GET /api/tickets/:id` – Detail tiket berdasarkan UUID
- `PATCH /api/tickets/:id` – Update tiket (transisi status, ubah prioritas, penugasan staf)
- `DELETE /api/tickets/:id` – Menghasilkan HTTP 405 Method Not Allowed (tiket dipertahankan untuk audit trail)

### Diskusi & Catatan Internal (Memerlukan JWT):
- `POST /api/tickets/:id/comments` – Tambah pesan publik atau catatan staf (`is_internal: true`)
- `GET /api/tickets/:id/comments` – Daftar riwayat komentar tiket

### Manajemen Pengguna (Admin Only, Memerlukan JWT + Role `admin`):
- `GET /api/users` – Daftar seluruh pengguna (mendukung query `page`, `limit`, `search`)
- `POST /api/users` – Buat pengguna baru secara administratif
- `PATCH /api/users/:id` – Update data pengguna (nama, email, role)
- `PATCH /api/users/:id/role` – Update role pengguna secara spesifik
- `DELETE /api/users/:id` – Hapus pengguna (dilindungi jika memiliki riwayat tiket/komentar atau akun admin terakhir)

---

## 11. Cara Menjalankan Database Migration

File migrasi database terletak di folder `backend/migrations/`.

### Migrasi Up (Membuat tabel & indeks):
```bash
cd backend
go run ./cmd/migrate -direction=up -path=./migrations
```

### Migrasi Down (Rollback tabel & indeks):
```bash
cd backend
go run ./cmd/migrate -direction=down -path=./migrations
```

### Seeding Akun Awal:
```bash
cd backend
go run ./cmd/seed
```

---

## 12. Cara Menjalankan Test

### Backend Test Suite:
Backend dilengkapi pengujian menyeluruh (unit tests, integration tests, concurrency tests, RBAC tests, state machine tests, dan JWT validation tests):

```bash
cd backend
go test -v -count=1 ./...
```

Hasil test mencakup:
- `internal/config`: Validasi secret production dan loader environment.
- `internal/middleware`: Validasi signature JWT, expiry token, role hierarchy, dan fail-closed DB verification.
- `internal/service`: Unit test state machine status tiket, kalkulasi paginasi, dan isolasi hak akses.
- `internal/repository`: Validasi query SQL dan parameterized escaping.
- `tests`:
  - `auth_test.go`: Alur login, register, profile, dan bcrypt.
  - `ticket_test.go`: 16 skenario pengujian siklus tiket, pencarian, filter, dan penolakan transisi terlarang.
  - `user_management_test.go` & `prompt5_test.go`: CRUD user admin dan penolakan demosi.
  - `user_concurrency_test.go`: Uji ketahanan konkurensi (20 goroutine paralel) untuk mencegah demosi admin terakhir.

### Frontend Build Test:
```bash
cd frontend
npm run build
```
Memastikan bundling Vite dan validasi sintaks Vue 3 berjalan sukses tanpa error.

---

## 13. Asumsi dan Batasan Implementasi

1. **Browser Client Communication:** Frontend berjalan sebagai Single Page Application di browser pengguna. Panggilan API diarahkan ke `http://localhost:8080/api` atau melalui Nginx reverse proxy `/api/`. Hostname internal docker seperti `backend:8080` tidak digunakan langsung di client browser.
2. **Penghapusan Tiket:** Sesuai standar audit trail enterprise, penghapusan tiket (`DELETE /api/tickets/:id`) mengembalikan HTTP 405 Method Not Allowed untuk menjaga riwayat operasional.
3. **Penghapusan Pengguna:** Akun pengguna yang sudah memiliki relasi tiket atau komentar dilindungi oleh integritas referensial basis data (`ON DELETE RESTRICT`) dan menghasilkan HTTP 409 Conflict untuk menjaga keutuhan riwayat helpdesk.
4. **Idempoten Seeding:** Seeder database dirancang idempoten (`ON CONFLICT DO NOTHING`) sehingga aman dijalankan berulang kali tanpa memicu duplikasi data.
5. **Session Invalidation:** Middleware backend memverifikasi status user ke database secara *real-time* pada setiap request, sehingga jika role diubah oleh admin atau user dihapus, hak akses segera terbarui tanpa perlu menunggu token JWT kedaluwarsa.
