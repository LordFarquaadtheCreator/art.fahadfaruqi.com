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
| `src/lib/utils/parallax.ts` | one shared loop for the plates' scroll drift |
| `src/lib/utils/cursor.svelte.ts` | the number the pointer carries over the grid |
| `src/lib/actions/count.ts` | animates a number to its new value; writes it outright when hidden |
| `src/lib/actions/reveal.ts` | entrance for what is below the fold; never claims what is on screen |
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
serving the previous HTML and, for hashed filenames, 404s or stale bytes. A file
copied into `build/` after startup 404s for the same reason — no restart, no index
entry. Symptom:
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
- **The layer draws over the photograph, never instead of it.** Each frame's `<img>` is
  always on screen and always loaded; the quad is composited on top of it. The layer
  therefore has no way to blank the gallery: if it cannot start, cannot fetch a plate's
  bytes, or loses its context, the effect is simply absent. `data-gl="live"` marks a
  plate the loop has drawn for, for diagnostics only — no CSS hides anything on it.
- The canvas itself is `visibility: hidden` until the loop has produced its first frame,
  so a layer that never draws cannot cover the page either.
- Textures are released for frames more than `MARGIN` (400 CSS px) outside the
  viewport, and `dropTexture` closes the `ImageBitmap`.
- `const EFFECTS` in `PlateLayer.ts` is the master switch for displacement and
  the RGB split. Turning it off leaves a pure pass-through, which is the state
  to debug a mis-registered quad in — a seam shows up immediately.
- `const GRADE` and `const GRAIN` are separate switches for the print: a lifted-black
  S-curve and warm halation around the highlights, then two octaves of animated grain
  weighted towards the midtones. They are separate because they fail differently — too
  much grade flattens a photograph, too much grain makes it look dirty. `GRAIN` is an
  amplitude, so `0` is off.
- Both arrive through `uGrade`/`uGrain` as *uniforms*, so the shader has one code path:
  switching either off mixes to the untouched photograph rather than compiling a
  different program.
- The page-wide grain plates stay, but their opacity dropped (`--grain-opacity: 0.16`
  dark, `0.055` light) — the photographs now carry real grain, and at the old strength
  the two multiplied into mud.
- The grade lives in the fragment shader for a reason: the CSS hover zoom it replaced
  could not be seen at all once the quad covered the image. Anything that changes how a
  photograph looks belongs here, not in a rule on `.plate__image`.
- **A shader change cannot be verified by building.** `npm run build` never compiles
  GLSL, so a broken shader ships silently and only fails in a browser that draws a
  frame — which is exactly what the automation tab never does. Compile the source in a
  page yourself (`gl.compileShader` + `getProgramInfoLog`) before believing it works.
  The vertex shader only compiles with three's injected declarations
  (`attribute vec3 position; attribute vec2 uv; uniform mat4 projectionMatrix; uniform
  mat4 modelViewMatrix;`) — without them the errors are about three, not about you.
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

- **Never let CSS hide content that only JavaScript can reveal.** The grid cells used
  to be `opacity: 0` with a `[data-revealed='true']` rule the reveal observer set — so
  any failure to hydrate, or a bundle that did not load, left every photograph invisible
  while all the text rendered. That is a mostly-blank page, and in the light theme it
  reads as a white one. The default state is now visible: the CSS hides a cell only when
  the action has explicitly claimed it (`[data-revealed='false']`), and a cell already in
  the viewport is never claimed at all.
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
- **State that JavaScript sets belongs in a bound class, not a `data-` attribute.**
  Svelte's compiler matches attribute selectors against the markup literally, so
  `.masthead[data-scrolled='true']` is reported as an unused selector (and stripped)
  when the markup says `data-scrolled="false"`. Declaring the attribute does not help
  — only an *equal* literal counts. A `class:` binding is compiler-visible and scoped,
  so the rule survives and still wins on specificity. An action-set attribute needs
  `:global()`, declared *after* the scoped base rule, because the two tie and source
  order decides.
- **A computed style read straight after a class or attribute change lies in this
  automation tab.** Transitions only advance when frames are produced, and a hidden tab
  produces none, so `getComputedStyle` reports the *starting* value of a rule that is
  in fact winning. This has now cost time twice (the light-theme probe, the masthead's
  scrolled state). Inject `*{transition:none!important}` first, then read.
- **The hidden automation tab delivers neither clicks nor scroll events.** CDP
  `Input.dispatchMouseEvent` moves the pointer (the cursor-index label updates, so
  pointer handlers are verifiable) but produces no `click` event at all — a listener
  added by hand counts 0. Drive interactions with DOM events instead:
  `el.dispatchEvent(new MouseEvent('click', {bubbles: true, composed: true}))` reaches
  Svelte's delegated handler, and `window.dispatchEvent(new Event('scroll'))` exercises
  a scroll handler (the masthead's `--progress` then reads 0.0857 at 1500px of 17501px,
  which is the arithmetic checked end to end). Anything genuinely routed through
  `requestAnimationFrame` — the parallax loop, a count-up — cannot be verified here.
- **`scroll-behavior: smooth` on `html` means `window.scrollTo` animates.** With no
  frames it never arrives, so probe scrolling with `{behavior: 'instant'}`.
- **Release virtual time when you are done with it.** `Emulation.setVirtualTimePolicy`
  with a `budget` leaves the clock paused; the next navigation then never completes and
  `document.documentElement` is `null`. Release with `{policy: 'advance'}` and no budget.

### Motion

Every animation on the site follows the same four rules.

- **An entrance never decides whether something is visible.** The hidden state exists
  only behind an attribute the script sets, or inside a `@keyframes` `from` frame. If
  the script dies, the page is unanimated, not empty.
- **Restarting a CSS animation needs a new name.** An animation does not re-run when
  its element is re-rendered; `GalleryGrid` toggles between `set-in-a` and `set-in-b`
  as `pass` alternates, which is how filtering re-plays the wipe without re-mounting a
  plate (and without throwing away decoded pixels or uploaded textures).
- **Transforms live on their own layer.** The scroll drift sits on `.cell__slide`, a
  wrapper inside `.cell`, so it never fights the cell's entrance transition over the
  same property.
- **Reduced motion is one global rule** (`src/app.css`): durations collapse to
  0.001ms **and `animation-iteration-count` is reset to 1**. The second half is not
  optional — the grain plates and the marquee run `infinite`, and collapsing only the
  duration turns them into thousands of cycles a second. JS-driven motion — the
  count-up, the parallax loop, the viewer's open/close — checks
  `prefers-reduced-motion` itself, and the count-up writes its final value outright
  when `document.hidden`, because a hidden document produces no frames to count with.
- **A marquee is two identical runs translated by half the track.** `.marquee__track`
  holds two `.marquee__run` elements and animates to `translate3d(-50%, 0, 0)`, so the
  loop is seamless because the second run lands exactly where the first began; hover
  holds it still. Anything less than half the track jumps at the seam.

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
- **One way to filter.** `SetIndex` in the masthead is the only navigation. The hero
  used to carry a second, larger copy of the same list — a set table with counts and
  dates — and it was removed: two controls doing the same job made the top of the page
  read as a contents page rather than as a photograph. The hero is the name and the
  three-line credit, then the band, then the work. There is no loading skeleton any
  more either; the hero's count line reports the fetch instead.
- **Layers.** `Atmosphere.svelte` renders two fixed layers: `.backdrop` (z-index 0)
  holding the pointer-tracked glow and the two vignettes, and `.overlay` (z-index 6)
  holding the two grain plates. The glow is a light source **behind** the page, so
  `main` and `.footer` are lifted with `position: relative; z-index: 1` — remove that
  and the page content drops behind the light. Over the photographs there are exactly
  two things: the plate canvas (z-index 4) and the grain plates (6).
- **Glow motion.** The pointer target is lerped at a very low follow factor (~0.022
  per frame) so the light trails and wobbles rather than tracking the cursor. It stays
  CSS: the WebGL pass this build does have draws the photographs, not the light.
- **Theme.** Dark by default, light available through the toggle, stored in
  `localStorage.theme`. `src/app.html` resolves it before first paint, so the toggle
  and that inline script must agree on the stored values. Palette, grain and glow
  strength are tokens in `src/app.css`. The two vignettes are the exception: a
  gradient's colours cannot interpolate, so both sit in `Atmosphere.svelte` as fixed
  layers that cross-fade by opacity when the theme changes.
- **Sticky set headers.** Each `.set__head` is sticky below the masthead and carries a
  pinned state, reported by a one-pixel `.set__sentinel` above it: the browser has no
  such state, so the sentinel's position carries it. Pinned, the header tightens and
  firms up, the way the masthead does once you leave the top.
- **Placeholders.** The grid blurs up from the 24px LQIP. The viewer does **not**: at
  full size a 24px source is an unrecognisable wash, so it stands the grid's own 800px
  derivative in behind the photograph — already fetched by the grid, so usually decoded
  — and fades it out when the 2200px file arrives.
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
  hidden tab, so the grid looks empty: cells that are below the fold stay at
  `opacity: 0` until they are observed. Force the loads and the reveal state:
  `document.querySelectorAll('.plate__image').forEach(i => { const s = i.getAttribute('src'); i.removeAttribute('loading'); i.src = s; })`
  and `document.querySelectorAll('.cell').forEach(c => { if (c.dataset.revealed === 'false') c.dataset.revealed = 'true'; })`
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
