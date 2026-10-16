// Runs against the disposable Compose stack in CI; never point at production.
import assert from "node:assert/strict";
const base = "http://localhost:8080";
let up = false;
for (let i = 0; i < 40; i++) {
  try {
    const r = await fetch(base + "/api/health");
    if (r.ok) {
      up = true;
      break;
    }
  } catch {}
  await new Promise((resolve) => setTimeout(resolve, 1000));
}
assert.ok(up, "Compose stack must become healthy");
for (const path of [
  "/",
  "/vehicles",
  "/dealer",
  "/login",
  "/register",
  "/dashboard",
]) {
  const response = await fetch(base + path);
  assert.equal(response.status, 200, path + " must render");
  const html = await response.text();
  assert.ok(html.includes("MICHEAL"), "Brand missing from " + path);
  assert.ok(html.includes("Chat with us"), "Chat widget missing from " + path);
}
const login = await fetch(base + "/api/auth/login", {
  method: "POST",
  headers: { Origin: base, "Content-Type": "application/json" },
  body: JSON.stringify({
    email: process.env.OWNER_EMAIL,
    password: process.env.OWNER_PASSWORD,
  }),
});
assert.equal(
  login.status,
  200,
  "Provisioned owner must be able to log in through gateway",
);
assert.equal((await login.json()).role, "owner");
const cookie = login.headers.get("set-cookie");
assert.ok(cookie.includes("HttpOnly"), "Session cookie must be HttpOnly");
const stats = await fetch(base + "/api/admin/stats", {
  headers: { Cookie: cookie.split(";")[0] },
});
assert.equal(stats.status, 200, "Authenticated owner statistics");
const denied = await fetch(base + "/api/auth/logout", {
  method: "POST",
  headers: {
    Origin: "https://untrusted.example",
    Cookie: cookie.split(";")[0],
  },
});
assert.equal(denied.status, 403, "Gateway must preserve Origin protection");
console.log(
  "Production Compose smoke checks passed: health, all pages, chat widget, administrator provisioning/login, authorized API and origin protection.",
);
