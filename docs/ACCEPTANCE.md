# Acceptance and validation

## Automated validation

- Go role/vehicle/identifier policy tests, origin rejection and limiter tests.
- PostgreSQL integration test: registration cannot supply an administrator role; customer cannot edit cars; manager cannot grant permissions; unpublished cars are hidden; publishing works; enquiries persist; customer cannot read another customer's conversation; owner gets a live WebSocket event; manager reply remains in history; permission revocation blocks access; corrupt uploads are rejected; valid image crosses Go → Python → public WebP; logout invalidates the session.
- Python resizing, metadata removal, unsupported/corrupt file and size rejection.
- Next.js production compilation and static route generation.
- Docker image builds and Compose startup in CI; production HTTP smoke checks cover public pages, the floating widget, one-time administrator setup/refusal, owner login, authenticated statistics and Origin rejection.

Local Go tests skip PostgreSQL integration unless `TEST_DATABASE_URL` is configured. A passing local run alone does not establish database integration success; inspect the CI backend job.

## Manual checks before launch

1. Run Compose and set up the two administrators; setup must refuse a second run.
2. Owner updates profile, photo, phone, email, address and WhatsApp number; verify the public dealer page.
3. Manager adds a draft car, uploads two images, reorders them, saves and publishes. Verify the cover, gallery, price, status and filters.
4. Archive a listing and confirm it disappears publicly. Check reserved/sold statuses.
5. Register two different customers. Each opens the bottom-right widget and sends a message; neither sees the other's messages.
6. Open owner and manager sessions in separate browser profiles. Reply from each, confirm team attribution, close/reopen a conversation, then sign out and sign back in to verify history.
7. Stop/restart the API between customer messages and staff viewing history; saved messages must still be present.
8. Submit a vehicle enquiry, update it to contacted/closed, and check the WhatsApp link uses the selected vehicle.
9. Owner revokes each manager permission. Confirm HTTP access fails even after reloading or using a previously open tab. Restore permissions as needed.
10. Check phone-sized and desktop layouts, keyboard navigation, readable focus states, form errors, dark mode persistence, empty states, and service-offline errors.
11. Verify HTTPS, Secure/HttpOnly cookie flags, invalid-Origin rejection, image type rejection, backups and restoration.

Browser interaction/visual review and a real hosting deployment must be completed in the target environment; a successful compile is not a substitute for these checks.
