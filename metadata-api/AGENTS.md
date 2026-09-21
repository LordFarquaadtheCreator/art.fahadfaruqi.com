# AGENTS.md - metadata-api

Cloudflare Worker that exposes metadata from the R2 `assets` bucket as JSON.

The directory and the Worker are both named `metadata-api`.

## Shape of the thing

- One source file: `src/index.js`. ES module export, no build step, no dependencies.
- Config: `wrangler.toml`. R2 bucket `assets` is bound as `ASSETS_BUCKET`.
- Route: `https://assets.fahadfaruqi.com/api/*`, zone `fahadfaruqi.com`.
- `workers_dev = false`, so there is no `*.workers.dev` URL. Use `wrangler dev`.
- This directory is plain files in the parent repo. Do not re-initialise a git
  repo here; the parent owns version control.

## Request handling

`fetch(request, env, ctx)` does two things:

1. Calls `route` inside a `try`/`catch`. **Any** throw becomes a
   `502 {"error":"Internal error"}` so that step 2 still runs.
2. Passes whatever came back through `finalizeResponse`.

Without the `try`/`catch`, a throw escapes as a bare `500` with no CORS headers,
which a cross-origin caller reports as a CORS failure rather than a server error.
Never return early from `fetch` without going through `finalizeResponse`.

`route(request, env, ctx)`:

- Computes `path` as `url.pathname.slice(4)` with trailing slashes stripped, so
  `/api/metadata` and `/api/metadata/` both yield `/metadata`.
- `OPTIONS` on **any** `/api/*` path returns a bare `204`. This happens before
  the cache lookup and before any R2 call, so a preflight cannot collide with a
  cached GET.
- `GET` or `HEAD` on `/metadata` goes to `handleMetadataRequest`.
- Everything else returns `404 {"error":"Not found"}`.

Do not rebind `request` inside `route`. The handler needs the real Request object
to build its cache key. An earlier revision reassigned `request` to the path
string, which silently broke both the method check and the cache key.

`handleMetadataRequest(request, env, ctx)`:

1. Build the cache key as `new Request(request.url)` — **always a GET**, whatever
   the caller sent. `cache.put` rejects any method but GET, and keying HEAD on the
   GET form means HEAD reads and fills the same entry as GET. The query string is
   part of the URL, so `?v=N` busts the cache.
2. `caches.default.match`; return the hit if present.
3. `listAllObjects(env)`, then respond with `{ count, objects }` plus the cache
   headers.
4. Hand the write to `ctx.waitUntil`. Not `await` — see below.

`listAllObjects(env)` pages `env.ASSETS_BUCKET.list({ cursor, limit: PAGE_SIZE })`
until `listing.truncated` is false, and maps each object to
`{ url, key, size, uploaded, etag }` with `obj.customMetadata` spread over it.
Past `MAX_OBJECTS` it throws rather than returning a partial gallery.

## Why `ctx.waitUntil`, not `await`

`await caches.default.put(...)` puts the cache write on the response's critical
path, so every miss pays the write latency before the client sees anything, and a
`put` rejection fails a request that already succeeded. `waitUntil` moves the
write off the response path and turns a failed write into a log line. Measured
with a 300ms write: awaited, the client waited 305ms; with `waitUntil`, under
100ms. Removing `ctx` from the signatures silently reintroduces this.

## Errors

- Anything thrown in `route` is logged and returned as
  `502 {"error":"Internal error"}` with CORS attached. The real error goes to
  `console.error`, never to the client.
- A bucket larger than `MAX_OBJECTS` is a hard failure on purpose. A silently
  truncated listing renders an incomplete gallery with no signal, which is worse
  than a clear 502.

## CORS

- `finalizeResponse` is the only place CORS headers are set. It always sets
  `Access-Control-Allow-Origin: *`.
- `Access-Control-Allow-Methods` (from `ALLOWED_METHODS`), `Allow-Headers`, and
  `Max-Age` are **preflight-only** and attached only when the method is `OPTIONS`.
  A GET or HEAD must not carry them. Keeping all three gated together matters —
  gating two and not the third is two rules for one concept.
- `Access-Control-Allow-Headers` echoes `Access-Control-Request-Headers`. It
  cannot be a wildcard.
- `ALLOWED_METHODS` must stay in sync with what `route` actually accepts.
- `finalizeResponse` builds a new `Response` instead of mutating headers, because
  responses from `caches.default.match` carry immutable header guards and `set()`
  throws. This is also where HEAD's body is nulled.
- The cached copy is deliberately CORS-free; headers are attached on every serve.
  Do not start caching the post-`finalizeResponse` response, or CORS gets baked
  into the entry and the echoed header freezes.

## Caching

- The Cache API entry lives and dies by `Cache-Control` on the response passed to
  `put()`, so `s-maxage=604800` governs the stored entry's TTL and `max-age`
  governs the browser. Both come from the one header.
- **The cache is per-data-center, not globally replicated.** Content cached in one
  data center does not exist in another until that one is asked. Do not describe
  this as a global edge cache.
- `cache.put` throws if the request method is not GET, if the response status is
  `206`, or if the response carries `Vary: *`.
- `stale-while-revalidate` and `stale-if-error` are not supported by the Cache API.
- Separately, the Cache API and the CDN cache are independent mechanisms. This
  Worker is cached through the Cache API only; there is no cache rule involved.

## Conventions

- R2 custom metadata keys become field names in `objects[]`. The spread happens
  last, so a metadata key named `url`, `key`, `size`, `uploaded`, or `etag`
  overwrites the generated value. Treat those as reserved.
- `uploaded` is `Date.prototype.toISOString()`; `etag` is R2's raw etag.
- TTLs and limits live in the constants at the top of the file: 7 days entry,
  1 day browser, 1 day preflight, `PAGE_SIZE`, `MAX_OBJECTS`.

## Gotchas

- The listing is cached for 7 days. Newly uploaded objects will not appear until
  that entry expires or the Worker is redeployed (which repopulates that data
  center on the next request). A `?v=N` query param bypasses it.
- The predecessor Worker `assets-api` is deleted and cannot be deleted again; the
  route `assets.fahadfaruqi.com/api/*` now belongs to this Worker. If you rename
  the Worker, delete the current one **before** deploying the rename. Only one
  Worker can hold a route, so two matching `assets.fahadfaruqi.com/api/*` is a
  conflict rather than a takeover.
- `url.pathname.slice(4)` assumes the `/api` prefix. That holds only because the
  route in `wrangler.toml` forwards `/api/*` and nothing else. Widen the route and
  this breaks silently.
- `.wrangler/` and `.dev.vars` are gitignored. `.dev.vars` is where `wrangler dev`
  secrets belong; keep it ignored.

## Deploy

Already deployed as `metadata-api`, owning the route `assets.fahadfaruqi.com/api/*`. To deploy, do:

```sh
wrangler deploy
```

Authentication is the part an agent usually trips on. If there are issues, stop and tell the user. Do not attempt anything yourself regarding auth.

## Verification

There is no test suite and no `package.json`. To exercise the whole routing matrix
without deploying, import the module and stub the globals it touches:
`globalThis.caches.default` (`match`/`put`), `env.ASSETS_BUCKET.list`, and an
`ctx` with `waitUntil` that collects the promises so you can await them yourself.
Then call `worker.fetch(new Request(url, init), env, ctx)`.

Cover, at minimum: `GET` returns `200` and caches once; a second `GET` from cache
still carries CORS; `HEAD` returns `200` with an empty body and reuses the GET
entry; `OPTIONS` returns `204` with `ALLOWED_METHODS` and an echoed
`Access-Control-Request-Headers`; `GET /api/metadata/` (trailing slash) returns
`200`; `GET /api/nope` and `POST /api/metadata` return `404` with CORS; and an R2
that throws returns `502` **with** CORS rather than rejecting.

```sh
wrangler dev   # curl localhost:8787/api/metadata, plus -I and -X OPTIONS on it
wrangler deploy
```

Verifying a deploy needs no credentials, since the endpoint is public. `curl` covers
the matrix against production:

```sh
curl -s https://assets.fahadfaruqi.com/api/metadata | jq '.count, .objects[0]'
curl -s -I https://assets.fahadfaruqi.com/api/metadata            # 200, empty body
curl -s -i -X OPTIONS -H 'Access-Control-Request-Headers: authorization' \
  https://assets.fahadfaruqi.com/api/metadata                     # 204, echoed header
curl -s -o /dev/null -w '%{http_code}\n' https://assets.fahadfaruqi.com/api/nope
```

Check `count` against the bucket and confirm a sample object's metadata fields. A
second `GET` coming back with `cf-cache-status: HIT` and a non-zero `age` is how you
confirm the Cache API write actually landed. A `HEAD` carrying the right
`content-length` with no body confirms the HEAD path, and a `GET` on a non-`/api` path
such as an image key should still be served by R2.
