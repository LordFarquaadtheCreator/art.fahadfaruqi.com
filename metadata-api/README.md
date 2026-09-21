# Assets Bucket Metadata API

A Cloudflare Worker that lists the objects in an R2 bucket and returns each one with its custom metadata as JSON.

Images themselves are served by R2 directly. This Worker only answers the metadata API.

- Metadata API: `https://assets.fahadfaruqi.com/api/metadata`
- Images: `https://assets.fahadfaruqi.com/<key>`

## Endpoints

| Method | Path | Response |
| --- | --- | --- |
| `GET` | `/api/metadata` | `{ count, objects[] }`, see below |
| `HEAD` | `/api/metadata` | Same headers as `GET`, empty body |
| `OPTIONS` | any `/api/*` | `204`, CORS preflight headers, empty body |
| other method, or unknown path | any `/api/*` | `404` with `{ "error": "Not found" }` |

If R2 fails, or the bucket holds more than the Worker's ceiling of 10,000 objects,
you get `502` with `{ "error": "Internal error" }` rather than a partially built
gallery.

Every other path on `assets.fahadfaruqi.com` is served by R2, not by this Worker.

## Response shape

```json
{
  "count": 1,
  "objects": [
    {
      "url": "https://assets.fahadfaruqi.com/paintings/sunset.jpg",
      "key": "paintings/sunset.jpg",
      "size": 1048576,
      "uploaded": "2026-09-21T12:00:00.000Z",
      "etag": "abc123",
      "title": "Sunset Over the Ocean",
      "medium": "Oil on Canvas",
      "year": "2024",
      "description": "A vibrant sunset over the Atlantic"
    }
  ]
}
```

`url`, `key`, `size`, `uploaded`, and `etag` are generated from R2. Everything
after that comes from the object's custom metadata, with the metadata keys used
as-is, so `title` in this example was uploaded as the `title` metadata header. An
object uploaded without any custom metadata returns just the five generated fields.

## Deploy

```sh
npm install -g wrangler
wrangler login
wrangler deploy
```

`wrangler.toml` binds the R2 bucket `assets` as `ASSETS_BUCKET` and routes
`assets.fahadfaruqi.com/api/*` on the `fahadfaruqi.com` zone, so deploying needs
access to that zone.

Two things to know if you deploy from a script or an agent rather than your own
terminal:

- `wrangler login` needs a browser and a prompt. Non-interactively, wrangler refuses
to run unless `CLOUDFLARE_API_TOKEN` is set in the environment.
- Only one Worker can hold a given route. If you ever rename the Worker, delete the
old one first; two Workers matching `assets.fahadfaruqi.com/api/*` is a conflict,
not a takeover.

## Local development

```sh
wrangler dev
curl http://localhost:8787/api/metadata
```

`workers_dev` is off, so there is no `*.workers.dev` URL. Test locally or hit
production.

## Upload images with metadata

Attach custom metadata at upload time with `x-amz-meta-*` headers:

```sh
wrangler r2 object put assets/paintings/sunset.jpg \
  --file=./sunset.jpg \
  --header="x-amz-meta-title:Sunset Over the Ocean" \
  --header="x-amz-meta-medium:Oil on Canvas" \
  --header="x-amz-meta-year:2024" \
  --header="x-amz-meta-description:A vibrant sunset over the Atlantic"
```

Each header becomes a field on that object in the API response. Avoid the
reserved names `url`, `key`, `size`, `uploaded`, and `etag`; metadata using those
overwrites the generated values.

## Fetch on the site

```js
fetch("https://assets.fahadfaruqi.com/api/metadata")
  .then((res) => res.json())
  .then((data) => {
    data.objects.forEach((art) => {
      console.log(art.title, art.url, art.medium, art.year);
    });
  });
```

CORS is open (`Access-Control-Allow-Origin: *`) on every response, including the
`404`s, so the browser can call it from any origin. No API key.

A plain GET is a "simple" request, so the browser sends it directly and no
preflight happens. The `OPTIONS` row in the table above exists for the day a
caller adds a custom header, sends `Content-Type: application/json`, or uses a
method other than GET/HEAD/POST. The Worker then answers the browser's automatic
preflight with `204`, echoes back whatever `Access-Control-Request-Headers` was
asked for, and caches the result for 1 day.

## Caching

A cache miss is answered immediately and backfilled in the background, so no client
ever waits on a cache write:

```mermaid
sequenceDiagram
    participant C as Client
    participant W as Worker
    participant K as Cache API
    C->>W: GET /api/metadata
    W->>K: match()
    K-->>W: miss
    W->>K: put() registered with waitUntil
    W-->>C: 200 response
    Note over W,K: invocation stays alive, client is done
    K-->>W: put settles
```

A hit returns at `match()` and the rest never runs.

| Layer | TTL |
| --- | --- |
| Workers cache (per data center) | 7 days |
| Browser | 1 day |

The response is cached for 7 days, keyed on the request URL. New uploads will not
show up until that entry expires or the Worker is redeployed. To see fresh data
immediately, add a query param (`?v=2`); it is part of the cache key, so it misses
the old entry.

Two things worth knowing about that 7 days:

- It is the stored response's TTL, taken from the `s-maxage` directive. `max-age`
covers the browser's copy.
- The cache is **per data center**. Content cached in one does not exist in another
until that one is asked, so a location serving its first request pays for a full
listing. There is no single global copy.

## Layout

```
metadata-api/
├── src/index.js     Worker source (single file)
├── wrangler.toml    R2 binding, route, compatibility date
├── AGENTS.md        Notes for agents/contributors working on this
└── README.md
```

No `package.json`, no build step, no tests. `wrangler deploy` is the whole
pipeline. This directory has no git repo of its own; the parent repository tracks
these files directly.
