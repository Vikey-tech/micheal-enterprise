# Deployment and operations

## Single-host deployment

1. Provision a Linux host with Docker and Compose, persistent disk, and a domain.
2. Clone the reviewed release and create `.env` with a strong random hexadecimal PostgreSQL password. Keep `.env` readable only by its operator (`chmod 600 .env`).
3. Configure `APP_ORIGIN=https://your-real-domain.example` without a trailing slash and `COOKIE_SECURE=true`.
4. Put a TLS reverse proxy on the host in front of `127.0.0.1:8080`. Forward WebSocket upgrades and preserve the incoming Origin header. Only expose HTTPS (and HTTP for redirects/certificate validation). Set HSTS at the TLS termination layer after verifying HTTPS works.
5. Run `docker compose up --build -d`, then the one-time account setup in the README. Avoid default/shared passwords. Remove all setup credential environment variables after use.
6. Sign in as owner, enter genuine dealership information, upload the owner photo, create real cars and publish listings. Confirm WhatsApp includes country code and digits only.
7. Complete the acceptance checklist with both staff accounts and two customer accounts before accepting real customers.

There is no hosted deployment included: domain, hosting credentials and business details must be supplied by the operator. This app is a single Go instance with PostgreSQL persistence. Do not scale API replicas without replacing the in-memory notification hub and rate limiter with shared infrastructure. Polling recovers missed notifications, but shared rate limiting is needed for multi-instance protection.

## Backups

Back up both the PostgreSQL database and the `uploads` volume on a schedule. Store encrypted backups off-host, restrict access, define retention, and test restoration.

```sh
docker compose exec -T db pg_dump -U micheal -d micheal -Fc > dealership.dump
```

The dump includes personal information and password hashes. Treat it as sensitive. Copy the uploads volume using your host's volume backup mechanism. Restore to an isolated environment first; verify cars, photos, staff access and conversation history. Do not use the integration-test database variable against a restored production database.

## Releases and monitoring

- Deploy a reviewed commit; never deploy unreviewed branches automatically.
- Health check: `GET /api/health` checks the database. Add external uptime monitoring and container log/health alerts.
- Logs contain operational errors rather than request bodies/passwords. Restrict log access and retention.
- Current schema initialization is idempotent for the initial schema. Future schema changes require explicit versioned migrations and a backup/rollback plan; do not rely on `CREATE TABLE IF NOT EXISTS` to update columns.
- Retain a prior image and backup before releases. Container downgrade does not reverse database schema changes.
- Session duration is 24 hours, and logout invalidates the server-side token. Permission changes apply on every HTTP request and before WebSocket message delivery.

## Account recovery

Email verification, self-service password reset, password-change screens, and MFA are not part of this MVP. For a lost account, a trusted host/database operator must verify the person, generate a bcrypt hash using a secure local tool, update only the verified user's password hash using a parameterized database client, and delete that user's sessions in the same transaction. Never send a stored hash or a password through public issues. Add a tested recovery flow before broad public promotion if the business needs unattended account recovery.

## Limits and privacy

- Inventory API returns up to 1,000 listings; inbox up to 500 conversations; history shows the latest 500 messages in chronological order. Older messages remain in PostgreSQL. Add cursor pagination before these limits affect use.
- No payments, financing, purchase checkout, email/SMS notifications or file attachments in chat. Replies appear in the app, not in email.
- No online/away indicator is inferred; the widget explains that messages are saved when the team is offline.
- Customer names, emails, enquiries and chat messages are stored. Publish the business's privacy/retention policy before launching and define an operator-assisted deletion/export process.
- Rate limiting uses the direct peer address and does not trust arbitrary forwarded headers. Behind nginx, public requests share a per-process limit. Tune gateway-level limits for expected traffic; add trusted client-IP handling and a shared limiter before higher traffic deployment.
- Orphaned uploads (removed photos, canceled edits) remain until operator cleanup. Do not delete images referenced by vehicle/profile JSON. Set disk usage alerts.
- Images are public once uploaded. Customer messages and enquiries are authenticated/permission protected.
- Content Security Policy permits inline Next.js bootstrap scripts and styles. A nonce-based policy is a future hardening option.
