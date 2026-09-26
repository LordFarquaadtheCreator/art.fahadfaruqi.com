# art.fahadfaruqi.com

Photographs by [Fahad Faruqi](https://fahadfaruqi.com) — a static gallery built with
SvelteKit and served from GitHub Pages. The photographs themselves live in Cloudflare
R2, served straight from the bucket for the images and through a small Worker for the
listing. Nothing about the gallery is baked into the build: the page loads, fetches the
metadata, and draws whatever the bucket contains.

**Live:** https://art.fahadfaruqi.com

## Why it is built this way

The masters are 6016×4016 WebP at quality 85, 1–8 MB each — they were PNGs of 110–140 MB
until they were re-encoded in place with their metadata carried over. Each one has four
pre-generated WebP derivatives, and the gallery only ever references those. Curated
captions and EXIF travel with each file in R2 as custom metadata, which keeps the
repository free of content: publishing a new photograph is an upload, not a commit.

```
R2 bucket "assets"
 ├── <name>.webp                    masters (never loaded by the page)
 └── d/{w2200,w1600,w800,lqip}/<name>.webp
      │
      ├── assets.fahadfaruqi.com/<key>            images
      └── assets.fahadfaruqi.com/api/metadata     listing + captions + EXIF
            │
            ▼
      this app → groups by set → renders the gallery
            │
            ▼
      GitHub Pages → art.fahadfaruqi.com
```

## The gallery

- **Offset grid.** A 12-column layout at desktop widths with explicit per-cell column
  starts, so the composition stays asymmetric and the negative space is deliberate.
  Below 1024px the plates fall into a single column.
- **Portrait-aware.** Plates whose photograph is taller than it is wide get a narrower
  span, so they sit inside the rhythm instead of towering over their neighbours.
- **Blur-up loading.** Each cell reserves its aspect ratio from a 24px placeholder, so
  the page does not shift as the 800px images arrive. The viewer stands the grid's own
  800px derivative behind the photograph instead, because at full size a 24px source is
  an unrecognisable wash.
- **Viewer.** Clicking a plate opens the full-screen viewer with the 2200px image, the
  caption, and the camera, lens, focal length, aperture, shutter, date and file size.
  Arrow keys move between photographs, Escape closes, and swipes work on touch.
- **Dark by default** with a light option; the choice is remembered. Palette, grain and
  light are CSS custom properties, and the theme is resolved before first paint.
- **A treated print.** Visible plates are drawn as textured quads on one WebGL canvas
  over the photograph — displacement, an RGB split, a print grade, halation and grain —
  with the `<img>` still underneath, so a browser without WebGL loses the effect and
  nothing else.
- **A backdrop and a moving light.** Grain, a faint mesh and two vignettes sit behind the
  content; in front of them, a warm light follows the pointer through a base blob and a
  stretch oval drawn between the pointer and where the light has reached, so it trails,
  wobbles, and pulls thin when the pointer moves fast.
- **A readout register.** The header carries the site's tabs, a scroll-progress hairline
  and the theme switch; the set filter counts each set; hovering a plate carries its
  number by the pointer; the viewer frames the photograph with its position in the run.
  Every number comes from the listing.

## Running it locally

```sh
bun install
bun run dev        # dev server
bun run check      # svelte-check
bun run build      # static build into build/
bun run preview    # serve the build
```

`.env` (gitignored) may override `VITE_METADATA_API` and `VITE_CDN_BASE`; when unset,
the production URLs are used, which is what CI does.

## Layout of the source

| Path | What it is |
| --- | --- |
| `src/routes/+page.svelte` | the page: the index read, the wait band, set filtering, the viewer's state |
| `src/routes/+layout.svelte` | the site chrome and the scroll state that drives the header |
| `src/routes/about/+page.svelte` | the About page — a stub with no content yet |
| `src/lib/utils/metadata.ts` | API types, fetch, mapping to the `Photo` shape |
| `src/lib/utils/exif.ts` | EXIF rationals → display strings (`7/2` → `f/3.5`) |
| `src/lib/utils/group-images.ts` | grouping into sets, ordering |
| `src/lib/utils/variants.ts` | derivative URL convention |
| `src/lib/webgl/` | the plate layer: one canvas, one quad per visible plate, the print shaders |
| `src/lib/components/Header.svelte` | tabs, theme switch, scroll-progress hairline |
| `src/lib/components/Hero.svelte` | the name, the credit lines and the count |
| `src/lib/components/GalleryGrid.svelte` | per-set sections, offset grid, reveal |
| `src/lib/components/SetCard.svelte` | a set's chapter card and its sticky header |
| `src/lib/components/PhotoPlate.svelte` | one photograph: placeholder, image, caption, note |
| `src/lib/components/Lightbox.svelte` | viewer with the panel, keyboard and swipe |
| `src/lib/components/MetadataDisplay.svelte` | the EXIF and file-size panel |
| `src/lib/components/SetIndex.svelte` | ALL / set / count filter row |
| `src/lib/components/SplitText.svelte` | segmented slice assembly for the display type |
| `src/lib/components/ThemeToggle.svelte` | dark / light switch |
| `src/lib/components/Footer.svelte` | the footer row and the copy-email button |
| `src/lib/components/ToTop.svelte` | the scroll-to-top control |
| `src/lib/components/Atmosphere.svelte` | backdrop (light, vignettes, noise, mesh) and the pointer's readout |
| `src/lib/components/ErrorPage.svelte` | the template every error page renders |
| `src/app.css` | Tailwind v4 entry, palette, type and layout tokens |
| `metadata-api/` | the Cloudflare Worker behind `/api/metadata` |
| `scripts/` | Go CLI for managing images and their metadata in the bucket |

## Images and metadata

Each master in the bucket needs four WebP siblings, which the app derives from the
key by convention:

| Key | Width | Used for |
| --- | --- | --- |
| `d/w2200/<name>.webp` | 2200 | viewer |
| `d/w1600/<name>.webp` | 1600 | 2× displays |
| `d/w800/<name>.webp` | 800 | grid |
| `d/lqip/<name>.webp` | 24 | blur-up placeholder |

Captions come from custom metadata on the master — `set`, `number`, `title`,
`alttext`, `description` — alongside the EXIF the camera wrote. The Worker returns both
and filters the `d/` prefix out of the listing, so derivatives never show up as
photographs. `metadata-api/README.md` documents the API, `scripts/README.md` the CLI
that writes that metadata, and `AGENTS.md` the details of working in this repo
(including how the derivatives are generated).

## Deploy

The site builds and publishes itself on every push to `main`
(`.github/workflows/deploy.yml` → GitHub Pages). The Worker deploys separately, and
only when it changes:

```sh
bun run worker:deploy
```

## Design credit

The layout follows the Stefan Vitasović portfolio case study published on Codrops
(2025). The performance is still the cheap one: every photograph in the gallery is a
plain `<img>` that a browser without WebGL renders on its own, and the layer above it is
an effect rather than the delivery mechanism. The print treatment and the pointer-tracked
light are this site's own, built on that reference rather than copied from it.
