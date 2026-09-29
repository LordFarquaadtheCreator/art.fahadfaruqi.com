// Assets Metadata API Worker
// Serves at: https://assets.fahadfaruqi.com/api/metadata
// R2 serves images directly at: https://assets.fahadfaruqi.com/<key>

const EDGE_CACHE_TTL = 1800; // 30 minutes in seconds
const BROWSER_CACHE_TTL = 600; // 10 minutes in seconds
const PREFLIGHT_MAX_AGE = 86400; // 1 day in seconds
const PAGE_SIZE = 1000; // R2 list page size
const MAX_OBJECTS = 10000; // refuse to build a partial listing past this
const ALLOWED_METHODS = "GET, HEAD, OPTIONS";
const CDN_BASE = "https://assets.fahadfaruqi.com";
const ART_PREFIX = "art/";
const MASTER_PREFIX = "art/master/";
const COMPRESSED_PREFIX = "art/compressed/";

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

function stemOf(key, prefix) {
  return key.slice(prefix.length).replace(/\.[^./]+$/, "");
}

function num(value) {
  const n = Number.parseInt(value, 10);
  return Number.isFinite(n) ? n : null;
}

async function listAllObjects(env) {
  const masters = new Map();
  const compressed = new Map();
  let cursor;
  let seen = 0;
  do {
    // Custom and HTTP metadata are omitted from list results unless requested
    // explicitly; without `include` the API would return five bare fields per object.
    const listing = await env.ASSETS_BUCKET.list({
      cursor,
      limit: PAGE_SIZE,
      prefix: ART_PREFIX,
      include: ["customMetadata", "httpMetadata"],
    });
    for (const obj of listing.objects) {
      seen += 1;
      if (seen > MAX_OBJECTS) {
        // Fail loudly rather than serve a silently incomplete gallery.
        throw new Error(
          `bucket holds more than MAX_OBJECTS (${MAX_OBJECTS}) under ${ART_PREFIX}; refusing to build a partial listing`,
        );
      }
      if (obj.key.startsWith(MASTER_PREFIX)) {
        masters.set(stemOf(obj.key, MASTER_PREFIX), obj);
      } else if (obj.key.startsWith(COMPRESSED_PREFIX)) {
        compressed.set(stemOf(obj.key, COMPRESSED_PREFIX), obj);
      }
    }
    cursor = listing.truncated ? listing.cursor : null;
  } while (cursor);

  const objects = [];
  const stems = [...new Set([...masters.keys(), ...compressed.keys()])].sort();

  for (const stem of stems) {
    const master = masters.get(stem);
    const comp = compressed.get(stem);
    const primary = master ?? comp;
    const curated = primary.customMetadata ?? {};
    const dims = (comp ?? master).customMetadata ?? {};

    objects.push({
      // The shape the deployed client reads: `key`, `url` and `size` name the master.
      url: master ? `${CDN_BASE}/${master.key}` : `${CDN_BASE}/${comp.key}`,
      key: `${stem}.webp`,
      size: master ? master.size : comp.size,
      etag: primary.etag,
      master_url: master ? `${CDN_BASE}/${master.key}` : null,
      master_size: master ? master.size : null,
      compressed_url: comp ? `${CDN_BASE}/${comp.key}` : null,
      compressed_size: comp ? comp.size : null,
      ...curated,
      uploaded: curated.uploaded ?? primary.uploaded.toISOString(),
      width: num(dims.width) ?? null,
      height: num(dims.height) ?? null,
    });
  }

  return objects;
}
