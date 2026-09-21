// Assets Metadata API Worker
// Serves at: https://assets.fahadfaruqi.com/api/metadata
// R2 serves images directly at: https://assets.fahadfaruqi.com/<key>

const EDGE_CACHE_TTL = 604800; // 7 days in seconds
const BROWSER_CACHE_TTL = 86400; // 1 day in seconds
const PREFLIGHT_MAX_AGE = 86400; // 1 day in seconds
const PAGE_SIZE = 1000; // R2 list page size
const MAX_OBJECTS = 10000; // refuse to build a partial listing past this
const ALLOWED_METHODS = "GET, HEAD, OPTIONS";
const CDN_BASE = "https://assets.fahadfaruqi.com";

const CORS_HEADERS = {
  "Access-Control-Allow-Origin": "*",
};

export default {
  async fetch(request, env, ctx) {
    let response;
    try {
      response = await route(request, env, ctx);
    } catch (err) {
      console.error("metadata-api:", err);
      response = json({ error: "Internal error" }, 502);
    }
    return finalizeResponse(response, request);
  },
};

async function route(request, env, ctx) {
  const url = new URL(request.url);
  const path = url.pathname.slice(4).replace(/\/+$/, "") || "/";
  const method = request.method;

  // Answer preflights for any /api/* path before touching the cache or R2.
  if (method === "OPTIONS") {
    return new Response(null, { status: 204 });
  }

  if (path === "/metadata" && (method === "GET" || method === "HEAD")) {
    return handleMetadataRequest(request, env, ctx);
  }

  return json({ error: "Not found" }, 404);
}

function json(body, status) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function finalizeResponse(response, request) {
  const headers = new Headers(response.headers);
  for (const [name, value] of Object.entries(CORS_HEADERS)) {
    headers.set(name, value);
  }

  // Preflight-only headers. A normal GET or HEAD never receives these.
  if (request.method === "OPTIONS") {
    headers.set("Access-Control-Allow-Methods", ALLOWED_METHODS);
    headers.set(
      "Access-Control-Allow-Headers",
      request.headers.get("Access-Control-Request-Headers") ?? "",
    );
    headers.set("Access-Control-Max-Age", String(PREFLIGHT_MAX_AGE));
  }

  // HEAD returns GET's headers with no body.
  const body = request.method === "HEAD" ? null : response.body;

  return new Response(body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

async function handleMetadataRequest(request, env, ctx) {
  const cacheKey = new Request(request.url);

  const cached = await caches.default.match(cacheKey);
  if (cached) {
    return cached;
  }

  const objects = await listAllObjects(env);
  const response = new Response(
    JSON.stringify({ count: objects.length, objects }),
    {
      headers: {
        "Content-Type": "application/json",
        "Cache-Control": `public, max-age=${BROWSER_CACHE_TTL}, s-maxage=${EDGE_CACHE_TTL}`,
      },
    },
  );

  // Cache the CORS-free copy; finalizeResponse decorates every serve.
  // waitUntil keeps the write off the response path so a slow or failing cache
  // write cannot delay or fail a request that already succeeded.
  ctx.waitUntil(caches.default.put(cacheKey, response.clone()));

  return response;
}

async function listAllObjects(env) {
  const objects = [];
  let cursor;
  do {
    const listing = await env.ASSETS_BUCKET.list({ cursor, limit: PAGE_SIZE });
    for (const obj of listing.objects) {
      if (objects.length >= MAX_OBJECTS) {
        // Fail loudly rather than serve a silently incomplete gallery.
        throw new Error(
          `bucket holds more than MAX_OBJECTS (${MAX_OBJECTS}); refusing to build a partial listing`,
        );
      }
      objects.push({
        url: `${CDN_BASE}/${obj.key}`,
        key: obj.key,
        size: obj.size,
        uploaded: obj.uploaded.toISOString(),
        etag: obj.etag,
        ...obj.customMetadata,
      });
    }
    cursor = listing.truncated ? listing.cursor : null;
  } while (cursor);

  return objects;
}
