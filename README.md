# art.fahadfaruqi.com

Photographs by [Fahad Faruqi](https://fahadfaruqi.com) — a static gallery built with
SvelteKit and served from GitHub Pages. The photographs themselves live in Cloudflare
R2, served straight from the bucket for the images and through a small Worker for the
listing. Nothing about the gallery is baked into the build: the page loads, fetches the
metadata, and draws whatever the bucket contains.

**Live:** https://art.fahadfaruqi.com

## Why it is built this way

The originals are 6016×4016 PNGs of 110–140 MB each, so they can never be what a
browser loads. Each one has three pre-generated WebP derivatives, and the gallery only
ever references those. Curated captions and EXIF travel with each file in R2 as custom
metadata, which keeps the repository free of content: publishing a new photograph is an
upload, not a commit.

```
R2 bucket "assets"
 ├── <name>.png                    originals (never loaded by the page)
 └── d/{w2200,w800,lqip}/<name>.webp
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
  the page does not shift as the 800px images arrive.
- **Viewer.** Clicking a plate opens the full-screen viewer with the 2200px image, the
  caption, and the camera, lens, focal length, aperture, shutter and date. Arrow keys
  move between photographs, Escape closes, and swipes work on touch.
- **Dark by default** with a light option; the choice is remembered. Palette, grain and
  light are CSS custom properties, and the theme is resolved before first paint.
- **Film grain and a moving light.** Two grain plates are drawn over the page as a film
  layer; behind it, a warm light follows the pointer with a heavy lag, so it trails and
  wobbles rather than tracking the cursor exactly.

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
| `src/routes/+page.svelte` | the page: fetch, set filtering, hero, gallery, footer |
| `src/lib/utils/metadata.ts` | API types, fetch, mapping to the `Photo` shape |
| `src/lib/utils/exif.ts` | EXIF rationals → display strings (`7/2` → `f/3.5`) |
| `src/lib/utils/group-images.ts` | grouping into sets, ordering |
| `src/lib/utils/variants.ts` | derivative URL convention |
| `src/lib/components/GalleryGrid.svelte` | per-set sections, offset grid, reveal |
| `src/lib/components/PhotoPlate.svelte` | one photograph: placeholder, image, caption |
| `src/lib/components/Lightbox.svelte` | viewer with the EXIF panel |
| `src/lib/components/SetIndex.svelte` | ALL / set / count navigation |
| `src/lib/components/SplitText.svelte` | per-character assembly for the display type |
| `src/lib/components/ThemeToggle.svelte` | dark / light switch |
| `src/lib/components/Atmosphere.svelte` | grain layer and pointer-tracked light |
| `src/app.css` | Tailwind v4 entry, palette, type and grid tokens |
| `metadata-api/` | the Cloudflare Worker behind `/api/metadata` |
| `scripts/` | Go CLI for managing images and their metadata in the bucket |
| `website-draft.md` | the design brief, including the reference this build follows |

## Images and metadata

Each original in the bucket needs three WebP siblings, which the app derives from the
key by convention:

| Key | Width | Used for |
| --- | --- | --- |
| `d/w2200/<name>.webp` | 2200 | viewer |
| `d/w800/<name>.webp` | 800 | grid |
| `d/lqip/<name>.webp` | 24 | blur-up placeholder |

Captions come from custom metadata on the original — `set`, `number`, `title`,
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
(2025), as recorded in `website-draft.md`. The grain and pointer-tracked light here are
CSS approximations of that site's WebGL layer.
