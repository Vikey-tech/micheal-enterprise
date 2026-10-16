# Architecture and API

Browser → nginx → Next.js frontend, or Go API (`/api`, `/uploads`). Go uses PostgreSQL and invokes a bounded Python image conversion subprocess. PostgreSQL and uploads are durable volumes; WebSocket notifications are transient hints, and the database is authoritative.

## Code boundaries

- `backend/cmd/server`: process configuration, database, graceful shutdown.
- `backend/cmd/setup`: one-time owner/manager provisioning.
- `backend/internal/models`: API/domain data types.
- `backend/internal/repository`: PostgreSQL connection, schema, session and chat/inventory queries.
- `backend/internal/service`: validation, role policy and cryptographic tokens.
- `backend/internal/middleware`: request identity, security headers, origin checks and limits.
- `backend/internal/handlers`: HTTP boundary/orchestration and endpoint-specific SQL; WebSocket subscriptions.
- `frontend/app`: Next.js routes; `frontend/components`: shared navigation, authentication, chat, API helpers and vehicle cards.
- `tools/optimize_images.py`: Pillow validation, bounds, orientation normalization, metadata removal and WebP conversion.

Vehicles/profile use validated JSONB documents; users, sessions, conversations, messages and enquiries use relational columns and foreign keys. Unique partial indexes ensure only one owner and one manager. Users cannot select their role during registration. No endpoint removes the owner, changes roles or allows a manager to grant permissions.

## HTTP API

JSON errors use `{ "error": "human-readable message" }`. Mutations require `Origin` exactly equal to `APP_ORIGIN`. Browser sessions use `me_session` cookies; only SHA-256 token digests are stored. Cookie tokens use 32 cryptographically random bytes. API responses are not cached.

| Endpoint | Access / purpose |
| --- | --- |
| `GET /api/health` | Database health |
| `POST /api/auth/register` | Customer only: name, email, password |
| `POST /api/auth/login` | Any role: email, password |
| `POST /api/auth/logout` | Invalidate current session |
| `GET /api/auth/me` | Signed-in identity and current permissions |
| `GET /api/vehicles` | Published vehicles |
| `GET /api/vehicles/{id}` | Published vehicle details |
| `GET /api/profile` | Public dealer profile |
| `POST /api/enquiries` | Vehicle ID, name, email, phone, body |
| `GET /api/admin/vehicles` | Vehicle permission; includes drafts/archive |
| `POST /api/admin/vehicles` | Vehicle permission; create |
| `PUT /api/admin/vehicles/{id}` | Vehicle permission; replace validated fields |
| `POST /api/admin/uploads` | Owner or vehicle permission; multipart `image` |
| `PUT /api/admin/profile` | Owner only |
| `GET /api/admin/stats` | Staff; counts constrained by permissions |
| `GET, PUT /api/admin/manager` | Owner only; vehicles/chat/enquiries permissions |
| `GET /api/admin/enquiries` | Enquiry permission |
| `PUT /api/admin/enquiries/{id}` | Enquiry permission; new/contacted/closed |
| `GET /api/conversations` | Own customer conversation or staff inbox with chat permission |
| `POST /api/conversations` | Customer only; create/reuse own dealership conversation |
| `PUT /api/conversations/{id}` | Chat permission; open/closed |
| `GET, POST /api/conversations/{id}/messages` | Conversation participant or permitted staff; POST body |
| `GET /api/ws` | Authenticated same-origin WebSocket; authorized message events |
| `GET /uploads/{name}` | Public generated WebP files only |

Chat sending uses HTTP POST with origin checks. A message and conversation update are committed in one transaction before notification. WebSockets deliver authorized message JSON and verify sessions/permissions before sending. Keepalive pings detect broken connections. Reconnecting clients fetch stored history; 10-second polling provides a fallback. Sending a new message reopens a closed conversation. Both staff identities display as members of the MICHEAL ENTERPRISE team.
