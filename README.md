# go-rest-boilerplate

Boilerplate REST API modular: Gin + GORM + PostgreSQL, dengan JWT auth
(access + refresh token) dan struktur RBAC (role & permission) siap pakai.

## Struktur

```
cmd/api/main.go          entrypoint, wiring & graceful shutdown
internal/config/         load .env / env vars
internal/database/       koneksi GORM + AutoMigrate
internal/models/         User, Role, Permission, RefreshToken
internal/repository/     data access layer (interface + implementasi GORM)
internal/service/        business logic (auth: register/login/refresh/logout)
internal/handler/        HTTP handler (gin)
internal/middleware/     JWTAuth, RequireRole, RequirePermission
internal/dto/            request/response struct + validasi
internal/router/         wiring route + dependency injection
internal/seeder/         seed role & permission default
pkg/jwt/                 generate & validate access token (JWT)
pkg/token/               generate & hash opaque refresh token
pkg/hash/                bcrypt password hashing
pkg/response/            format response JSON konsisten
```

## Auth flow

- **Access token**: JWT (HS256), umur pendek (default 5m), berisi `user_id`
  dan `roles`. Divalidasi tanpa hit DB → cepat.
- **Refresh token**: string random opaque (bukan JWT), disimpan di DB dalam
  bentuk hash SHA-256 (bukan raw), umur panjang (default 7 hari). Setiap kali
  dipakai untuk refresh, token lama **direvoke** dan token baru diterbitkan
  (rotation) — mencegah replay jika refresh token bocor.
- **Logout**: revoke refresh token yang dikirim, access token lama tetap
  valid sampai expired (karena JWT stateless) — kalau butuh revoke instan,
  perpendek `JWT_ACCESS_TTL`.

## RBAC (role & permission)

Skema: `users` ↔ `roles` (many2many via `user_roles`) ↔ `permissions`
(many2many via `role_permissions`). Sudah diseed 2 role default: `admin`
(semua permission `user:*`) dan `user` (tanpa permission, role default saat
register).

Dua middleware tersedia, pakai sesuai kebutuhan:

- `middleware.RequireRole("admin")` — cek role dari klaim JWT, tanpa query
  DB. Cocok untuk pengecekan kasar/cepat.
- `middleware.RequirePermission(rbacRepo, "user:delete")` — query DB setiap
  request, sehingga permission yang baru dicabut langsung berlaku (tidak
  menunggu access token expired). Cocok untuk aksi sensitif.

Menambah permission/role baru: tambahkan di `internal/seeder/seeder.go`,
lalu assign ke role lewat GORM association — tidak perlu ubah kode
middleware atau handler yang sudah ada.

## Menjalankan

### Opsi A — Docker (tidak perlu install PostgreSQL di lokal)

```bash
cp .env.example .env      # sesuaikan JWT_SECRET, kredensial DB bebas
make up                   # docker compose up -d --build
make logs                 # lihat log API
make down                 # stop semua service
```

Service yang jalan:

| Service    | Port lokal | Keterangan                              |
|------------|-----------|-------------------------------------------|
| `api`      | 8080      | REST API (Go)                             |
| `postgres` | 5432      | Database, data persisten via named volume |
| `adminer`  | 8081      | UI ringan untuk lihat isi database         |

`api` menunggu `postgres` lolos healthcheck (`depends_on: condition:
service_healthy`) sebelum start, dan `database.Connect` di kode juga retry
koneksi (10x, jeda 3 detik) untuk jaga-jaga kalau Postgres masih warm-up.
Di dalam compose network, `DB_HOST` di-override otomatis ke `postgres`
(nama service), meskipun `.env` kamu isi `localhost`.

Login ke Adminer di `http://localhost:8081`: System = PostgreSQL, Server =
`postgres`, isi user/password/database sesuai `.env`.

### Opsi B — Local Go (perlu PostgreSQL terpasang sendiri)

```bash
cp .env.example .env      # sesuaikan kredensial & JWT_SECRET
go mod tidy
make run                  # atau: go run ./cmd/api
```

Pastikan PostgreSQL sudah jalan dan database (`DB_NAME`) sudah dibuat.
Skema dibuat otomatis lewat GORM `AutoMigrate` saat startup.

## Endpoint

| Method | Path                        | Auth              | Keterangan                     |
|--------|-----------------------------|-------------------|---------------------------------|
| POST   | /api/v1/auth/register       | -                 | Daftar user baru (role `user`) |
| POST   | /api/v1/auth/login          | -                 | Login, dapat access+refresh    |
| POST   | /api/v1/auth/refresh        | -                 | Tukar refresh token → pair baru|
| POST   | /api/v1/auth/logout         | -                 | Revoke refresh token           |
| GET    | /api/v1/me                  | Bearer access     | Profil user login              |
| GET    | /api/v1/admin/ping          | Bearer + role     | Contoh `RequireRole`           |
| DELETE | /api/v1/admin/users/:id     | Bearer + permission | Contoh `RequirePermission`   |
