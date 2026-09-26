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
  ├── <name>.webp                 masters, 6016×4016 q85, 1-8 MB, never loaded by the page
  └── d/{w2200,w1600,w800,lqip}/<name>.webp   derivatives, the only images the page loads
        │
        ├── https://assets.fahadfaruqi.com/<key>            images, straight from R2
        └── https://assets.fahadfaruqi.com/api/metadata     Worker → { count, objects[] }
              │
              ▼
        this app (client-side fetch on load) → group by `set` → order by `number`
              │
              ▼
        bun run build → build/ → GitHub Pages (art.fahadfaruqi.com)
```

There is no server-side component in this repo's own build: the page is static
markup, and the gallery appears once the browser fetches the metadata API.

## Repository map

| Path | What it is |
| --- | --- |
| `src/routes/+page.svelte` | the page: fetch, set filtering, hero, gallery, footer |
| `src/routes/+layout.svelte` | font import, favicon, global CSS entry |
| `src/routes/+layout.ts` | `prerender = true`, `trailingSlash = 'never'` |
| `src/routes/+error.svelte` | the error boundary: one line, rendering `ErrorPage` with nothing passed |
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
| `src/lib/components/Atmosphere.svelte` | background blob layer + foreground grain |
| `src/lib/components/ErrorPage.svelte` | the error template: status, copy, the address that missed, the way back — what every `+error.svelte` renders, and what a failed index read draws with `status = 500` |
| `src/app.css` | Tailwind v4 entry, palette and layout tokens |
| `src/app.html` | pre-paint theme resolution; must stay in step with the toggle |
| `static/` | `favicon.png`, `robots.txt`, `.nojekyll` — copied verbatim into `build/` |
| `metadata-api/` | the Worker that serves `/api/metadata` |
| `scripts/` | Go CLI for managing the bucket. **Go only** — see Rules |
| `website-draft.md` | the design brief this build follows, including the reference |

## Local development

```sh
bun install
bun run dev        # vite dev server
bun run check      # svelte-check: keep this at 0 errors / 0 warnings
bun run build      # static build into build/
bun run preview    # serve build/
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
- `uploaded` is R2's own timestamp, so any re-upload resets it, and the gallery reads
  it as the set's date. `create --inherit <ext>` writes the original's value into
  custom metadata, and because custom metadata is spread over the generated fields
  that value wins.
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

Masters are 6016×4016 WebP at quality 85, 1–8 MB each. Nothing in the app loads them.
Every master is expected to have four WebP siblings, and `src/lib/utils/variants.ts`
derives the URL from the master's key by stripping the extension. The masters were PNGs
of 110–140 MB until they were re-encoded in place — same key, new extension — with their
metadata carried over by `create --inherit`, which is what keeps the key and its
derivatives aligned:

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
for v in w2200 w1600 w800 lqip; do
  curl -s -o /dev/null -w "d/$v/<name>.webp %{http_code}\n" \
    "https://assets.fahadfaruqi.com/d/$v/<name>.webp"
done
```

**There is no committed generator.** The script used to create these was deliberately
kept out of the repo (the user's rule: `scripts/` holds Go only). It lives at
`~/.hermes/cache/scratch/derivatives.mjs` on the machine that ran it, uses `sharp`,
reads the metadata API, skips derivatives that already return 200, and uploads through
`bunx wrangler r2 object put --remote` with
`cache-control: public, max-age=31536000, immutable`. If it is gone, rewriting it is
a small job: read the listing, resize to the four widths above, upload to the four
`d/` prefixes. It needs R2 credentials, which are in `scripts/config.yaml` (gitignored).
Its source bytes come from each object's `url`, which is now the WebP master, so a re-run
is one more lossy generation: regenerate from a fresh export rather than repeatedly.

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
- **A shader change cannot be verified by building.** `bun run build` never compiles
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
- **A DOM click before hydration does nothing.** The plates are in the prerendered HTML
  before Svelte attaches its handlers, so a probe that clicks the moment `.plate` appears
  can test nothing at all and report a false negative. Wait for a signal only script can
  produce — a class it sets, a count-up's final value — not for the element.
- **`* { transition: none !important }` does not reach pseudo-elements.** Reading a
  `::after` transform with only that rule in place reports the untouched start value in a
  hidden tab, which looks exactly like a broken selector. Kill `*::before` and `*::after`
  too before believing a pseudo-element's computed value.
- **Script execution can be switched off to see the pre-hydration page.**
  `Emulation.setScriptExecutionDisabled(true)` then a reload renders the prerendered markup
  with the stylesheets applied and no hydration at all — the only way to inspect the states
  that exist before the app boots, which is where the loading band lives and where the
  canvas colour is decided. Re-enable it and reload when done.
- **To count WebGL contexts, instrument `getContext` before the app boots.** Surplus
  contexts usually come from detached probe canvases, which no DOM query can see —
  `document.querySelectorAll('canvas')` shows the one canvas the app draws into while the
  browser is holding dozens. Install `Page.addScriptToEvaluateOnNewDocument` with a wrapper
  around `HTMLCanvasElement.prototype.getContext` that records the type and whether the
  canvas is attached, reload, then read the array back. That is what turned "one canvas, so
  one context" into the measured "27 plates, 30 contexts, 27 of them probes".
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
  holding the pointer-tracked blob and the two vignettes, and `.overlay` (z-index 6)
  holding the two grain plates. The blob is a light source **behind** the page, so
  `main` and `.footer` are lifted with `position: relative; z-index: 1` — remove that
  and the page content drops behind the light. Over the photographs there are exactly
  two things: the plate canvas (z-index 4) and the grain plates (6).
- **blob motion.** The pointer target is lerped at a very low follow factor (~0.022
  per frame) so the light trails and wobbles rather than tracking the cursor. It stays
  CSS: the WebGL pass this build does have draws the photographs, not the light.
- **Theme.** Dark by default, light available through the toggle, stored in
  `localStorage.theme`. `src/app.html` resolves it before first paint, so the toggle
  and that inline script must agree on the stored values. Palette, grain and blob
  strength are tokens in `src/app.css`. The two vignettes are the exception: a
  gradient's colours cannot interpolate, so both sit in `Atmosphere.svelte` as fixed
  layers that cross-fade by opacity when the theme changes.

### One context, not one per plate

The layer is built on a single canvas, and for a long time the page was quietly creating
more than twenty. `supported()` — the "can this browser give us a WebGL context at all"
check — ran on **every** plate registration, and each run left a detached probe canvas
holding a context the browser would not collect for a while. Measured on a loaded page:
**27 plates, 30 contexts, 27 of them probes.**

Chrome caps live contexts per page and kills the oldest once you pass the cap — the
console warning `Too many active WebGL contexts. Oldest context will be lost.`, followed by
`THREE.WebGLRenderer: Context Lost.` A canvas whose context has been taken from it
composites as an **opaque white rectangle**. This one is `position: fixed; inset: 0` at
`z-index: 4` — above every photograph, below the masthead — so what a visitor saw was a
white page with the nav bar intact. That is the white flash. Nothing about it is
dev-specific: it happened in production too.

Three changes, each sufficient to stop it alone:

- **Ask once per document.** `probeResult` memoises the answer, and the probe's context is
  handed straight back through `WEBGL_lose_context`. Measured after: **27 plates, 2
  contexts** — one probe, one layer.
- **Release the context on teardown.** `WebGLRenderer.dispose()` frees three's own
  resources but leaves the context to the collector, so `dispose()` also calls
  `forceContextLoss()`.
- **Take the canvas out of the page when a context goes.** `handleContextLost` deletes
  `data-ready`, which is what `.plate-canvas` is shown by, so a lost context can never
  paint over the page. The plates are plain `<img>` elements; losing the layer costs the
  effect and nothing else.

Verified by forcing the loss with `WEBGL_lose_context.loseContext()`: with `data-ready`
set, the canvas reports `visibility: visible` at z-index 4; after the loss `data-ready` is
gone and the computed visibility is `hidden`.

### The inline canvas

The canvas is painted before the stylesheets are. `<html>` carries an inline
  `background` and `color-scheme: dark`, `app.html`'s script repaints both for a
  light-theme visitor, and `theme-color` covers the browser chrome. Without that the
  document has no background at all until the CSS lands, and the browser fills the gap
  with its own canvas — which follows the *used* color-scheme, so it is white on a machine
  set to light mode. Note this is a real gap but it is **not** the flash that gets
  reported: what people actually saw was the WebGL layer losing its context, above. The
  literals duplicate `--bg`: nothing can read a custom property that early. `ThemeToggle`
  repaints the same surfaces on toggle, the meta included.

### The instrument pass

A second register, drawn from the Death Stranding interface: amber, and the readouts that
give the amber something to report.

- **Amber is for what is live.** `--accent` marks the active set's marker, the masthead's
  progress hairline, the pointer's index readout, the plate caption's hover rule, the
  focused element, the chapter card's rule, and the viewer's status and position. Body
  text, captions, set names, the EXIF keys and the footer stay monochrome. That restraint
  is what keeps this compatible with the editorial layout, and it is the reference's own
  rule: cold amber, for active and *projected* state, never for warmth.
- **One amber cannot serve both themes.** `#dc8d18` is 7.42:1 on the dark ground and
  2.36:1 on the light one — failing even the 3.0 bar for UI components. Light therefore
  uses `#8a4a0a` (6.06:1): same hue, different lightness. Measure before changing either.
- **Two warm colours are one too many.** The blob's stops sat at hue 22 and 13 while the
  accent is 35.8 — which read as two colours the moment the accent existed. They are now
  36 and 31, so the light and the signal belong to one family.
- **The readout face ships.** `--font-mono` is IBM Plex Mono (latin subset, 400 only,
  14.7 KB of woff2) rather than a system stack, so a readout looks the same on every
  machine instead of rendering as whatever the visitor's OS calls monospace. Captions are
  capitalised with their units attached.
- **Every number is real.** The viewer's `PLATE 07 / 27` is the position in the current
  list; the panel's File row comes from the listing's own `size`; a chapter card's date is
  the set's most recent upload. An `<img>` reports no byte progress, so the viewer says
  `DECODING` and shows no percentage — there is nothing honest to build one from.
- **Framing stays in the margin.** The viewer's brackets are inset by the layout's own
  padding and verified not to overlap the photograph. That is the line between an
  instrument and a drawing over somebody's work.
- **The chapter card is additive.** Each set opens with a band carrying its number, name,
  count and date, replaying on every entry by toggling a class — an animation only
  restarts when it is re-applied. The sticky header still carries the same facts, so
  nothing depends on the card rendering.
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

### The plate's note

Every photograph carries a description from the listing, and it arrives on the plate
itself after a dwell — `DWELL_MS` in `PhotoPlate.svelte`, two seconds.

- **Two ways to dwell, decided by the device.** Hover for anything with a pointer, and an
  `IntersectionObserver` at 0.65 visibility for anything without one; `(hover: hover) and
  (pointer: fine)` picks the branch, and the observer is not even constructed on a desktop.
  Verified on a touch-emulated load: one 0.65 observer per plate, 27 of them.
- **Keyboard skips the wait.** A visitor who has tabbed to a plate has already chosen it,
  so the note arrives at once. Its focus state is component state — `plate__media--focused`
  — because `:focus-visible` cannot be verified while the page is not the focused window,
  and a selector that cannot be tested is a selector that will break silently.
- **The note lands on the photograph, not in the caption.** The caption is a grid item, so
  growing it would push every row beneath it down and shift the page under somebody who is
  reading. Measured: the figure's height and the next plate's `top` are identical before
  and after the reveal.
- **Nothing is behind the note.** The caption still carries the title and the viewer
  carries the description too, so a plate whose note never arrives is still fully
  readable — this is an addition, never the only path to the text.

### The wait

The index arrives over the wire, so the page opens with a state rather than with
photographs. It is a band in the site's own register: a readout naming where the index
comes from, an indeterminate amber hairline, and four `--bg-elev` frames holding the
grid's place.

- **Nothing is invented.** The frames are empty on purpose — no placeholder metadata, no
  photograph. The hairline travels rather than fills while the wait lasts, and when the read
  lands it is drawn once across the full width: the read is over, no percentage is claimed.
- **It hands over; it is not cut.** The panel being replaced keeps its node, takes
  `class:leaving` in the same flush that the next state takes the room, and leaves the flow in
  that frame — so no frame paints it in flow beside the arriving grid, and nothing below it
  moves. Its parts then leave in the order they were there for: the hairline drawn once, the
  report put away, the frames folded to their top edge, the panel clear and dropped by
  `DISSOLVE_MS`. It is `inert` and `pointer-events: none`, so the gallery beneath is live from
  the first frame, and reduced motion gets no ghost at all. Svelte's own `transition:` cannot
  do this job — an outro's keyframes are applied only after its dummy animation's finish
  event, so the outgoing element stays in flow for a frame or two, which is the jump this
  exists to avoid.
- **The band reports; it does not decorate.** Its readout is the state of the read
  (`Receiving index` / `Index received` / `Index unavailable`), and the dot holds instead
  of pulsing once there is nothing left to wait on. What the read returned belongs to the
  hero's line, which counts it out as the band leaves.
- **A read that fails is the error page.** The index is read in the browser, so its failure
  belongs to the page: the failure view draws the same `ErrorPage` the router uses, with
  `status = 500`, the reason in the readout register and another attempt as the action. It
  cannot be thrown into `+error.svelte` — only a `load` that throws reaches the error
  boundary, and the read cannot live in one: a universal load is re-run at hydration, which
  awaits it, so the prerendered band would be hydrated against a gallery.
- **The layer draws over the departure.** The plate canvas is fixed at z-index 4, above
  everything inside `main` (z-index 1), so in WebGL mode arriving plates composite over the
  band wherever they overlap. The strip and the hairline sit in the margin above the grid,
  which is where the departure is read.
- **It is announced, not just drawn:** `role="status"` and `aria-live="polite"`.

## Deployment

Two independent targets.

**The site** — `.github/workflows/deploy.yml` runs on push to `main`: `bun install`,
`bun run build`, `actions/upload-pages-artifact`, `actions/deploy-pages`. Pages is
configured with `build_type: workflow` and the custom domain
`art.fahadfaruqi.com`, so the repository root is not what gets served: the artifact's
own `build/index.html` is. Nothing outside `build/` reaches the live site — anything
placed at the repository root would be invisible on the web, `robots.txt` and
`favicon.png` included (they come from `static/`). Steps:

```sh
bun run check && bun run build        # catch it locally first
git push origin main
gh run list --limit 1 --json databaseId --jq '.[0].databaseId' | xargs -I{} gh run watch {} --exit-status
```

**Poll for the run that matches the commit you pushed.** A poll issued the instant after the
push can still see the *previous* commit's run and report *its* green — which is a
confirmation of the wrong thing. Match the `headSha` before believing it:

```sh
gh run list --limit 40 --json headSha,conclusion \
  --jq '.[] | select(.headSha|startswith("<short-sha>"))'
```

**The Worker** — separate, manual, and only needed when `metadata-api/` changes:

```sh
bun run worker:deploy        # cd metadata-api && wrangler deploy
```

**The 404 page** — `svelte.config.js` passes `fallback: '404.html'` to `adapter-static`,
which writes `build/404.html` as an app shell. GitHub Pages serves that file for any path
that is not a file in the deploy, with a 404 status (confirmed: `custom_404: false`, and
the Pages default only applies when no `404.html` exists) — without it Pages shows its own
generic page. The shell carries no route payload, so the router resolves the address that
was actually asked for, finds no route, and renders the root error page with status 404 and
the message `Not Found`, which is what `src/routes/+error.svelte` reads. A prerendered
page would not do: it hydrates against the route it was built for, not against the
address it is served at.

Because `page.url` is the address that was asked for, `ErrorPage.svelte` can name it: the
404's subtitle is the address that missed, with the status as the display type above it,
which is the only place the visitor can see that, and the only clue to which link is broken.

The cost of the fallback is that the 404's markup arrives with the bundle rather than in
the file — it is the one page on the site that is empty without script. Everything about
how it looks is still the site's own, because the error page renders inside the same
layout: `ErrorPage.svelte` draws with the tokens in `src/app.css`, and the `Atmosphere`
layers are already mounted around it, rather than the second copy of the palette the old
`static/404.html` carried. Verify both halves from production:

```sh
curl -s -o /dev/null -w "%{http_code}\n" https://art.fahadfaruqi.com/no-such-path   # 404
curl -s https://art.fahadfaruqi.com/no-such-path | grep -c '_app/immutable'         # >0
```

## Verification

Against production, no credentials needed:

```sh
curl -s https://art.fahadfaruqi.com/ | grep -o '<link rel="icon"[^>]*>'
curl -s "https://assets.fahadfaruqi.com/api/metadata?cb=1" \
  | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d["count"]); print(sorted(d["objects"][0]))'
for v in w2200 w1600 w800 lqip; do curl -s -o /dev/null -w "$v %{http_code}\n" \
  "https://assets.fahadfaruqi.com/d/$v/<name>.webp"; done
```

In a browser, what "the gallery works" means: 27 `figure` elements across 3 sets, the
grid images loading from `d/w800`, the backdrop blob present behind the content, two
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
4. **Keep `bun run check` clean.** It currently reports 0 errors and 0 warnings.
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
