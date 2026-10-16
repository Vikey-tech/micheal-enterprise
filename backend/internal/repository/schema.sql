CREATE TABLE IF NOT EXISTS users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL, email text NOT NULL UNIQUE,
 password_hash text NOT NULL, role text NOT NULL CHECK(role IN ('owner','manager','customer')),
 permissions jsonb NOT NULL DEFAULT '{"vehicles":true,"chat":true,"enquiries":true}', created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS one_owner ON users(role) WHERE role='owner';
CREATE UNIQUE INDEX IF NOT EXISTS one_manager ON users(role) WHERE role='manager';
CREATE TABLE IF NOT EXISTS sessions (token_hash text PRIMARY KEY, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, expires_at timestamptz NOT NULL);
CREATE INDEX IF NOT EXISTS session_expiry ON sessions(expires_at);
CREATE TABLE IF NOT EXISTS vehicles (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), data jsonb NOT NULL, created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS profile (id integer PRIMARY KEY CHECK(id=1), data jsonb NOT NULL);
INSERT INTO profile(id,data) VALUES(1,'{"name":"MICHEAL ENTERPRISE","bio":"Personal service. Exceptional vehicles. Speak with our team to find your next car.","phone":"","whatsapp":"","email":"","address":"","image":""}') ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS conversations (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), customer_id uuid NOT NULL UNIQUE REFERENCES users(id), status text NOT NULL DEFAULT 'open' CHECK(status IN ('open','closed')), updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS messages (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), conversation_id uuid NOT NULL REFERENCES conversations(id), sender_id uuid NOT NULL REFERENCES users(id), body text NOT NULL CHECK(length(body) BETWEEN 1 AND 4000), created_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX IF NOT EXISTS messages_history ON messages(conversation_id,created_at);
CREATE TABLE IF NOT EXISTS enquiries (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), vehicle_id uuid REFERENCES vehicles(id), name text NOT NULL,email text NOT NULL, phone text NOT NULL, body text NOT NULL, status text NOT NULL DEFAULT 'new' CHECK(status IN ('new','contacted','closed')),created_at timestamptz NOT NULL DEFAULT now());
