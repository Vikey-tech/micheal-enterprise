# MICHEAL ENTERPRISE

A dealership MVP built with **Go, Next.js/React (JavaScript), Tailwind CSS, PostgreSQL, and Python/Pillow**. Prices use Nigerian naira. The repository starts with no invented vehicles, administrator passwords, or business contact details.

## Features

- Responsive milk/beige/chocolate storefront; persistent light/dark theme.
- Search/filter inventory by make, availability, price, year, transmission and fuel; price/year sorting.
- Vehicle pages with photo galleries, enquiries and configured WhatsApp contact.
- Customer registration/login, persistent conversation history, floating chat on every page.
- WebSocket reply notifications plus reconnect and periodic database refresh. Messages persist while staff are offline.
- Separate owner and manager accounts, protected by server-side permissions. No public admin registration or role-changing endpoint.
- Create/edit/publish/archive listings; available/reserved/sold status; featured cars; image ordering.
- Owner profile/contact editing, enquiries, inbox, statistics, and owner-controlled manager permissions.
- Python validates and converts uploaded JPEG/PNG/WebP files to bounded WebP images, removing metadata. Go invokes it with a timeout; no separate Python service.
- Opaque database sessions, bcrypt passwords, HttpOnly cookies, origin checks, per-process request limits, validated uploads, database constraints, non-root application containers.

## Run locally with Docker

Install Docker Engine with Compose v2. From the repository root:

```sh
cp .env.example .env
# Edit .env: replace POSTGRES_PASSWORD with a random hex secret.
docker compose up --build -d
```

Open **http://localhost:8080**. The gateway binds only to loopback. PostgreSQL and the application services have no exposed host ports. Database and uploads use named volumes.

### Create the two administrators

In an interactive Bash terminal, run the following. Passwords are read silently and are not put in command arguments or repository files. Use different, unique passwords of 12–72 bytes. Email addresses must differ.

```bash
read -r -p 'Owner name: ' OWNER_NAME
read -r -p 'Owner email: ' OWNER_EMAIL
read -r -s -p 'Owner password: ' OWNER_PASSWORD; printf '\n'
read -r -p 'Manager name: ' MANAGER_NAME
read -r -p 'Manager email: ' MANAGER_EMAIL
read -r -s -p 'Manager password: ' MANAGER_PASSWORD; printf '\n'
export OWNER_NAME OWNER_EMAIL OWNER_PASSWORD MANAGER_NAME MANAGER_EMAIL MANAGER_PASSWORD
docker compose run --rm --no-deps \
  -e OWNER_NAME -e OWNER_EMAIL -e OWNER_PASSWORD \
  -e MANAGER_NAME -e MANAGER_EMAIL -e MANAGER_PASSWORD api /app/setup
unset OWNER_NAME OWNER_EMAIL OWNER_PASSWORD MANAGER_NAME MANAGER_EMAIL MANAGER_PASSWORD
```

Setup is transactional, protected by a database advisory lock, and refuses to run if an administrator already exists. It is an operator CLI, never a web route. Sign in at `/login`, then use the dashboard to enter real contact details, add photos and vehicles, and publish listings. Uploaded images become public URLs; upload only material intended for public display.

## Development without Docker

Requirements: Go 1.26+, Node 22+, Python 3.12+, PostgreSQL 17.

```sh
python3 -m pip install -r tools/requirements.txt
cd frontend && npm ci
```

Run the API from `backend/`, with `DATABASE_URL`, `APP_ORIGIN=http://localhost:3000`, `COOKIE_SECURE=false`, `UPLOAD_DIR` set to an absolute writable directory, and `IMAGE_OPTIMIZER` set to the absolute path to `tools/optimize_images.py`. Then run `go run ./cmd/server`. Run `npm run dev` from `frontend/`. Next proxies `/api` and `/uploads` to `http://localhost:8081`. Production uses nginx for reliable WebSocket upgrades.

## Checks

```sh
cd backend
go test -race ./...
go vet ./...
cd ..
python3 -m unittest discover -s tools -v
cd frontend
npm ci
npm run build
```

`TEST_DATABASE_URL` enables the PostgreSQL integration test. **It truncates application tables: use a dedicated disposable database only.** GitHub Actions provisions one and exercises registration, admin/customer access, draft visibility, publishing, enquiries, conversation isolation, WebSocket delivery, saved history, permission revocation, logout, and Go-to-Python image uploads. CI also builds the application containers.

See [deployment](docs/DEPLOYMENT.md), [architecture and API](docs/ARCHITECTURE.md), and [acceptance checklist](docs/ACCEPTANCE.md).
