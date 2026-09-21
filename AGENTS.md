# AGENTS.md

Working notes for this repository. This file covers the front-end app, the build
and the deployment. `metadata-api/AGENTS.md` is the authoritative document for the
Worker, and `scripts/README.md` for the Go CLI.

## What this is

A photo portfolio at https://art.fahadfaruqi.com. SvelteKit, prerendered to static
files, published to GitHub Pages. The photographs and their metadata are not in the
repo: they live in the Cloudflare R2 bucket `assets`, which is read directly for
images and through a Worker for the listing.

## Architecture

```
Cloudflare R2 bucket "assets"
  ├── <name>.png                  originals, 110-140 MB, never loaded by the gallery
  └── d/{w2200,w800,lqip}/<name>.webp   derivatives, the only images the page loads
        │
        ├── https://assets.fahadfaruqi.com/<key>            images, straight from R2
        └── https://assets.fahadfaruqi.com/api/metadata     Worker → { count, objects[] }
              │
              ▼
        this app (client-side fetch on load) → group by `set` → order by `number`
              │
              ▼
        npm run build → build/ → GitHub Pages (art.fahadfaruqi.com)
```

There is no server-side component in this repo's own build: the page is static
markup, and the gallery appears once the browser fetches the metadata API.

## Repository map

| Path | What it is |
| --- | --- |
| `src/routes/+page.svelte` | the page: fetch, set filtering, hero, gallery, footer |
| `src/routes/+layout.svelte` | font import, favicon, global CSS entry |
| `src/routes/+layout.ts` | `prerender = true`, `trailingSlash = 'never'` |
| `src/lib/utils/metadata.ts` | API types, fetch, mapping to the `Photo` shape |
| `src/lib/utils/exif.ts` | EXIF rationals → display strings (`7/2` → `f/3.5`), date formatting |
| `src/lib/utils/group-images.ts` | grouping into sets, slugify, ordering |
| `src/lib/utils/variants.ts` | derivative URL convention |
| `src/lib/webgl/layer.ts` | module-level singleton + the Svelte action that registers a plate |
| `src/lib/webgl/PlateLayer.ts` | the canvas: one quad per visible plate, rect-driven, texture budget |
| `src/lib/webgl/shaders.ts` | GLSL for the plate quads |
| `src/lib/components/GalleryGrid.svelte` | per-set sections, offset 12-column grid, reveal observer |
| `src/lib/components/PhotoPlate.svelte` | one photograph: LQIP, image, caption |
| `src/lib/components/Lightbox.svelte` | full-screen viewer: EXIF panel, keyboard, swipe |
| `src/lib/components/SetIndex.svelte` | the ALL / set / count navigation |
| `src/lib/components/SplitText.svelte` | per-character assembly used for the display type |
| `src/lib/components/ThemeToggle.svelte` | dark/light switch |
| `src/lib/components/Atmosphere.svelte` | background glow layer + foreground grain |
| `src/app.css` | Tailwind v4 entry, palette and layout tokens |
| `src/app.html` | pre-paint theme resolution; must stay in step with the toggle |
| `static/` | `404.html`, `favicon.png`, `robots.txt`, `.nojekyll` — copied verbatim into `build/` |
| `metadata-api/` | the Worker that serves `/api/metadata` |
| `scripts/` | Go CLI for managing the bucket. **Go only** — see Rules |
| `website-draft.md` | the design brief this build follows, including the reference |

## Local development

```sh
npm install
npm run dev        # vite dev server
npm run check      # svelte-check: keep this at 0 errors / 0 warnings
npm run build      # static build into build/
npm run preview    # serve build/
```

`.env` (gitignored, so absent in CI) may set `VITE_METADATA_API` and
`VITE_CDN_BASE`. When they are unset — which is the case for every CI build — the
code uses the production URLs compiled in at `src/lib/utils/metadata.ts` and
`src/lib/utils/variants.ts`. **Changing the API host therefore means editing those
defaults too**, not just `.env`; a local-only change to `.env` will not reach the
deployed site.

### Pitfall: `vite preview` caches its file index

`vite preview` indexes `build/` when it starts. Rebuilding while it runs leaves it
serving the previous HTML and, for hashed filenames, 404s or stale bytes. Symptom:
a module you just built is served with a different hash or a 9-byte body. Fix:
restart the preview process after every rebuild. Restart it with a pattern that
cannot match the shell you are restarting from — `pkill -f '[v]ite preview'`, not
`pkill -f 'vite preview'`: the latter matches the chained command's own arguments and
kills the wrapper shell before the new server starts. Always confirm what is actually
being served before drawing conclusions from the browser:

```sh
f=$(grep -o '_app/immutable/nodes/[^"]*\.js' build/index.html | tail -1)
shasum -a 256 < "build/$f"                      # on disk
curl -s "http://localhost:4173/$f" | shasum -a 256   # what the server returns
```

## The metadata contract

`GET https://assets.fahadfaruqi.com/api/metadata` returns `{ count, objects[] }`,
one entry per original. Fields are **flat**, not nested, and R2 lowercases the
custom-metadata keys, so the names in the response are:

| Group | Fields |
| --- | --- |
| Generated by the Worker | `url`, `key`, `size`, `uploaded`, `etag` |
| Curated at upload time | `set`, `number`, `title`, `alttext`, `description` |
| EXIF, carried through as strings | `make`, `model`, `lens`, `focallength`, `fnumber`, `exposuretime`, `isospeedratings`, `datetimeoriginal` |

Consequences to keep in mind:

- Custom metadata is spread over the generated fields, so the five generated names
  are effectively reserved — a metadata key named `url` or `etag` overwrites it.
- `fnumber` and `focallength` arrive as rationals (`7/2`, `18/1`), `exposuretime`
  as `1/60`, `datetimeoriginal` as an ISO string with offset. `src/lib/utils/exif.ts`
  is the only place that should format these.
- `isospeedratings` can be an empty string.
- The worker requests `include: ["customMetadata", "httpMetadata"]` from R2's
  `list()`. R2 omits both unless asked, and dropping that option silently strips
  every curated field and all EXIF from the response — this was the original bug.
- The listing is cached at the edge for 7 days. New uploads do not appear until the
  entry expires or the Worker is redeployed. `?cb=<anything>` busts it, which is how
  the derivative generator and anyone debugging should read it.
- `d/` keys are filtered out of the listing (`DERIVATIVE_PREFIX` in the Worker).
  Adding a new derivative variant does not require touching the gallery.

## Image derivatives

Originals are 6016×4016 PNGs at 110–140 MB. Nothing in the app loads them. Every
original is expected to have four WebP siblings, and `src/lib/utils/variants.ts`
derives the URL from the original key by stripping the extension:

| Key | Width | Quality | Used for |
| --- | --- | --- | --- |
| `d/w2200/<name>.webp` | 2200 | 82 | lightbox / viewer |
| `d/w1600/<name>.webp` | 1600 | 80 | the 1600w srcset candidate — 2× displays and the WebGL quads |
| `d/w800/<name>.webp` | 800 | 78 | gallery grid cells at 1× |
| `d/lqip/<name>.webp` | 24 | 60 | blur-up placeholder; also carries `naturalWidth`/`naturalHeight`, which the grid uses to reserve each cell's aspect ratio |

`PhotoPlate.svelte` offers the three real sizes through `srcset` with
`sizes="(min-width: 1024px) 55vw, 100vw"`, so the browser picks: 800 for a 1×
desktop cell, 1600 for the same cell on a 2× display or a full-width phone.

A missing derivative is a broken image in the gallery with no runtime error, because
the URL is built by convention rather than looked up. After any upload, verify:

```sh
for v in w2200 w800 lqip; do
  curl -s -o /dev/null -w "d/$v/<name>.webp %{http_code}\n" \
    "https://assets.fahadfaruqi.com/d/$v/<name>.webp"
done
```

**There is no committed generator.** The script used to create these was deliberately
kept out of the repo (the user's rule: `scripts/` holds Go only). It lives at
`~/.hermes/cache/scratch/derivatives.mjs` on the machine that ran it, uses `sharp`,
reads the metadata API, skips derivatives that already return 200, and uploads through
`node_modules/.bin/wrangler r2 object put --remote` with
`cache-control: public, max-age=31536000, immutable`. If it is gone, rewriting it is
a small job: read the listing, resize to the four widths above, upload to the four
`d/` prefixes. It needs R2 credentials, which are in `scripts/config.yaml` (gitignored).

## The WebGL layer

`src/lib/webgl/` draws every visible plate as a textured quad on **one** fixed
canvas (`.plate-canvas`, `z-index: 4`). One context, not one per photograph —
27 contexts would be 27 copies of the GPU state and a hard browser limit.

- `layer.ts` owns the module-level singleton and the Svelte action
  `registerPlate(frame)` that `PhotoPlate.svelte` applies to `.plate__frame`.
  The action returns a teardown, so the layer ref-counts and disposes itself
  when the last plate unmounts.
- Every tick reads each frame's `getBoundingClientRect()` and copies it onto the
  plane. The layer knows nothing about layout, so the offset grid, the portrait
  spans and any future CSS change are picked up for free.
- **The `<img>` is the photograph, not a fallback.** It is hidden
  (`data-gl="live"`) only after its quad has actually drawn pixels, and it is
  never gated on CORS: the layer fetches its own bytes (`fetch` →
  `createImageBitmap` → `THREE.Texture`), so if CORS breaks, or the bytes are
  unreachable, the plate simply stays a plain image instead of blanking.
- Textures are released for frames more than `MARGIN` (400 CSS px) outside the
  viewport, and `dropTexture` closes the `ImageBitmap`.
- `const EFFECTS` in `PlateLayer.ts` is the master switch for displacement and
  the RGB split. Turning it off leaves a pure pass-through, which is the state
  to debug a mis-registered quad in — a seam shows up immediately.
- `data-webgl="off"` on `<html>` stops the loop and hands every photograph back
  to its own `<img>`. The CSS half of that lives in `src/app.css` and needs
  `!important` because the rule it overrides is scoped (`.plate__image.svelte-hash`).
- A lost GL context does the same thing by itself: the loop stops, textures are
  dropped, the DOM images come back. It must never keep running — continuing
  would re-upload onto a dead canvas while the images stay hidden.

### Bucket CORS is a prerequisite

The layer needs the bucket to answer CORS for the origins the site runs on.
This is bucket state, not repo state, so it is not in git:

```sh
cat > /tmp/r2-cors.json <<'JSON'
{"rules":[{"allowed":{"origins":["https://art.fahadfaruqi.com","https://lordfarquaadthecreator.github.io","http://localhost:4173","http://localhost:5200"],"methods":["GET","HEAD"],"headers":["*"]},"exposeHeaders":["ETag","Content-Length"],"maxAgeSeconds":86400}]}
JSON
cd metadata-api && npx wrangler r2 bucket cors set assets --file /tmp/r2-cors.json
curl -s -D- -o /dev/null -H "Origin: https://art.fahadfaruqi.com" \
  https://assets.fahadfaruqi.com/d/w800/sam-10.webp | grep -i access-control-allow-origin
```

Note the shape: `wrangler r2 bucket cors set` takes the Wrangler
`{rules:[{allowed:{…}}]}` form, **not** the R2 REST `AllowedOrigins` shape, which
it rejects. Only the Worker's `/api/*` responses set CORS by themselves.

### Traps that cost time here

- **Never alias a package to itself.** `svelte.config.js` briefly carried
  `kit: { alias: { three: 'three' } }`, added by mistake in the redesign commit.
  Nothing imported `three` then, so it sat there until the layer did — and the
  build failed with `[UNLOADABLE_DEPENDENCY] Could not load three`, which reads
  like a resolver bug rather than a config one. `resolve.alias` in
  `vite.config.ts` cannot override `kit.alias`, so that fix has to be made in
  `svelte.config.js`.
- **Do not set `texture.colorSpace = SRGBColorSpace` on a pass-through shader.**
  It makes three use the `SRGB8_ALPHA8` internal format, the GPU decodes to
  linear on sample, and the photograph comes out visibly darker than the same
  file drawn by the browser. Measured in this page: a 128 grey samples as **55**
  (73 levels down). Leave the default and write the sampled value straight out.
- **Keep three's default mipmap filters when downscaling.** The grid draws a
  1600px source into a ~790px quad; `minFilter = LinearFilter` with no mipmaps
  under-samples and reads as soft. `Texture`'s defaults
  (`LinearMipmapLinearFilter` + `generateMipmaps`) are the correct chain.
- **Cached derivatives predate CORS.** The `d/` objects are served
  `immutable, max-age=31536000`. A browser that fetched one before the bucket
  allowed this origin holds a copy that can never satisfy a CORS request, and
  `cache: 'default'` will keep failing on it. That is why `fetchBitmap` retries
  once with `cache: 'reload'` — only a network refetch clears it.

## Design system

The reference for the layout is the Stefan Vitasović portfolio (2025) as recorded in
`website-draft.md`. What that means in code:

- **Grid.** 12 columns at ≥1024px with explicit `grid-column` starts per cell index
  (`.cell--0` … `.cell--7`), which is what makes the composition asymmetric and
  leaves intentional holes. Below 1024px every plate is one full-width column.
- **Portrait handling.** A portrait photograph in a wide span would tower over its
  neighbours, so `GalleryGrid` matches it with
  `:has(:global(.plate--portrait))` and gives it a 4-column span, alternating columns
  by `:nth-child(even)`. `PhotoPlate` sets that class from the LQIP's aspect ratio.
  In the single-column layout, `max-width: calc(var(--ratio) * 74vh)` stops a portrait
  from being several screens tall.
- **Layers.** `Atmosphere.svelte` renders two fixed layers: `.backdrop` (z-index 0)
  holding the pointer-tracked glow and the vignette, and `.overlay` (z-index 6)
  holding the two grain plates. The glow is a light source **behind** the page, so
  `main` and `.footer` are lifted with `position: relative; z-index: 1` — remove that
  and the page content drops behind the light. Grain is the only thing drawn over the
  photographs.
- **Glow motion.** The pointer target is lerped at a very low follow factor (~0.022
  per frame) so the light trails and wobbles rather than tracking the cursor. It is
  an approximation of the reference's WebGL layer, not a port of it; a real WebGL
  pass was explicitly out of scope for this build.
- **Theme.** Dark by default, light available through the toggle, stored in
  `localStorage.theme`. `src/app.html` resolves it before first paint, so the toggle
  and that inline script must agree on the stored values. Palette, grain, glow and
  vignette strength are tokens in `src/app.css`; nothing hard-codes a colour.
- **Caption metadata.** The EXIF line under each plate is always visible at `--faint`
  and lifts to `--muted` on hover/focus. The lightbox shows the full panel.

## Deployment

Two independent targets.

**The site** — `.github/workflows/deploy.yml` runs on push to `main`: `npm ci`,
`npm run build`, `actions/upload-pages-artifact`, `actions/deploy-pages`. Pages is
configured with `build_type: workflow` and the custom domain
`art.fahadfaruqi.com`, so the repository root is not what gets served: the artifact's
own `build/index.html` is. Nothing outside `build/` reaches the live site — anything
placed at the repository root would be invisible on the web, `robots.txt` and
`favicon.png` included (they come from `static/`). Steps:

```sh
npm run check && npx vite build        # catch it locally first
git push origin main
gh run list --limit 1 --json databaseId --jq '.[0].databaseId' | xargs -I{} gh run watch {} --exit-status
```

**The Worker** — separate, manual, and only needed when `metadata-api/` changes:

```sh
npm run worker:deploy        # cd metadata-api && wrangler deploy
```

**The 404 page** — `static/404.html` is a standalone page, copied into the artifact,
which GitHub Pages serves for any path that is not a file in the deploy, with a 404
status (confirmed: `custom_404: false`, and the Pages default only applies when no
`404.html` exists). It deliberately avoids hashed asset names and JavaScript so it
renders even if the bundle fails, which means its design tokens are a copy of
`src/app.css` rather than a reference to it — update both if the palette moves.
`adapter-static` never emits a `404.html` (it only writes a `fallback` when one is
configured), so this file is the only 404 handling on the site.

## Verification

Against production, no credentials needed:

```sh
curl -s https://art.fahadfaruqi.com/ | grep -o '<link rel="icon"[^>]*>'
curl -s "https://assets.fahadfaruqi.com/api/metadata?cb=1" \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d["count"]); print(sorted(d["objects"][0]))'
for v in w2200 w800 lqip; do curl -s -o /dev/null -w "$v %{http_code}\n" \
  "https://assets.fahadfaruqi.com/d/$v/<name>.webp"; done
```

In a browser, what "the gallery works" means: 27 `figure` elements across 3 sets, the
grid images loading from `d/w800`, the backdrop glow present behind the content, two
grain layers in the overlay, and clicking a plate opening the viewer with its EXIF
panel populated. `metadata-api/AGENTS.md` lists the Worker's own matrix (GET, HEAD,
OPTIONS, trailing slash, 404s, CORS on errors).

### Pitfalls when verifying in a headless or background tab

- `loading="lazy"` images and the `IntersectionObserver` reveal do not fire in a
  hidden tab, so the grid looks empty and cells stay at `opacity: 0`. To check the
  markup and URLs anyway, force the loads:
  `document.querySelectorAll('.plate__image').forEach(i => { const s = i.src; i.removeAttribute('loading'); i.src = s; })`
- The layout breakpoint is 1024px. A small automation window (≈900px) silently
  renders the single-column layout — set the viewport to 1440×900 first
  (`Emulation.setDeviceMetricsOverride`) or you will conclude the offset grid is
  broken when it is not.
- Screenshots come back as JPEG. Fine grain does not survive the compression, so do
  not judge grain visibility from one; read the computed opacity and blend mode of
  `.grain--coarse` instead.
- **A hidden window produces no frames at all.** If the OS window is backgrounded,
  every tab reports `visibilityState: "hidden"` and `requestAnimationFrame` never
  fires — so the WebGL layer never draws, `data-gl` never appears and the plates
  legitimately stay as plain images. `Page.bringToFront`, `Page.startScreencast` and
  `Emulation.setVirtualTimePolicy` do not fix this. Verify the layer in a visible
  window; in a hidden one, use the `data-webgl="off"` flag to A/B the CSS half, and
  `document.querySelectorAll('.plate__frame[data-gl=live]').length` to see whether the
  loop has run at all.
- **`Emulation.setVirtualTimePolicy` freezes the page.** Setting it to `pause` and
  leaving it stops the clock, so the tab sits at `readyState: "loading"` with no
  `<body>` for as long as it is set — which reads exactly like a broken deploy. Release
  it with `{"policy":"advance","budget":N}` before concluding anything.
- A CSS transition reads as its *current animated value* in `getComputedStyle`, so in a
  frame-starved tab `opacity` never reaches its target and an override looks broken.
  Set `el.style.transition = 'none'` before measuring.

## Credentials and secrets

- `scripts/config.yaml` holds the Cloudflare API token and R2 keys in plain text. It
  is gitignored, along with the compiled `scripts/manage-images` binary. Keep it that
  way; this repository is public.
- `.env` is gitignored as well. Never commit either file, and never print their
  values into a log, a commit message or a chat.
- Wrangler authenticates from `~/.wrangler/config/default.toml`. If authentication
  fails, stop and tell the user; do not try to work around it.

## Rules for changes

1. **`scripts/` is Go only.** It exists to manage the metadata API's data. Throwaway
   scripts belong outside the repo, not in this directory.
2. **Never commit generated output** (`build/`, `.svelte-kit/`, `node_modules/`,
   `scripts/manage-images`, `scripts/config.yaml`, `.env`).
3. **No silent fallbacks.** If a fetch fails or a field is missing, the failure should
   be visible rather than papered over. The two `??` defaults for the API and CDN URLs
   predate this build and exist so CI can build without a `.env`.
4. **Keep `npm run check` clean.** It currently reports 0 errors and 0 warnings.
5. **Curated metadata is the content.** `title`, `alttext`, `description`, `set` and
   `number` come from the bucket, not from this repo; do not invent placeholder
   content in components.

## Leftovers and open items

- `package.json` still carries dependencies from the original scaffold that nothing
  imports: `three`, `@types/three`, `threlte`, `svelte-lightbox`, `svelte-bricks`,
  `@humanspeak/svelte-motion`. They are dead weight — several MB of install — and were
  left in place rather than removed unasked. `postcss`/`autoprefixer` are still used by
  `postcss.config.js`; keep those.
- `static/.assetsignore` is a leftover from the abandoned `adapter-cloudflare` setup
  (it lists `_worker.js` and `_routes.json`). Harmless, but meaningless now.
- The listing cache means a freshly uploaded photograph may take up to 7 days to
  appear. Redeploying the Worker is the quick way to force it.
- `https_enforced` is false on the Pages site, so plain `http://art.fahadfaruqi.com`
  is not redirected to HTTPS. It is a one-line API change if that is wanted.
