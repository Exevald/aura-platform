# shortlink

Product service for user-owned short links.

- `POST /api/shortlink/v1/links` creates a random 10-character short code for an absolute HTTP(S) URL.
- `GET /api/shortlink/v1/links` lists links owned by the authenticated user.
- `DELETE /api/shortlink/v1/links/{linkID}` deletes an owned link.
- `GET /s/{code}` is public and responds with `302 Found` or `404 Not Found`.

The management API is protected by the gateway with module ID
`d5b2367d-af92-4fa8-bc65-93b41f37a397`; the public redirect route bypasses session
and module access checks.

```shell
mise run //modules/shortlink:test
mise run //modules/shortlink:build
mise run //modules/shortlink:integration-tests
```
