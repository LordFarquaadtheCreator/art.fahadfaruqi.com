# art.fahadfaruqi.com

Photo portfolio. SvelteKit, statically prerendered, deployed to GitHub Pages; the
images and their metadata live in Cloudflare R2 behind a Worker.

## How it fits together

```
R2 bucket `assets`  ──►  Worker (metadata-api/)  ──►  this app  ──►  GitHub Pages
originals + derivatives   /api/metadata listing        fetch + render
```

- `https://assets.fahadfaruqi.com/<key>` — images, served straight from R2.
- `https://assets.fahadfaruqi.com/api/metadata` — every object with the custom
  metadata attached at upload time (`title`, `altText`, `description`, `set`,
  `number`) plus the EXIF values. See `metadata-api/README.md`.
- The app fetches that listing client-side on load, groups photos into sets by the
  `set` field, orders them by `number`, and renders the gallery.

## Running it

```sh
npm install
npm run dev        # vite dev server
npm run build      # static build into build/
npm run preview    # serve the built site
npm run check      # svelte-check
```

`.env` holds `VITE_METADATA_API` and `VITE_CDN_BASE`; both fall back to the
production URLs when unset, so a build without `.env` still works.

## Layout of the source

| Path | What it is |
| --- | --- |
| `src/routes/+page.svelte` | the whole page: fetch, filter state, hero, set index, footer |
| `src/lib/utils/metadata.ts` | API types, fetch, mapping to the `Photo` shape |
| `src/lib/utils/exif.ts` | raw EXIF rationals → display strings (`7/2` → `f/3.5`) |
| `src/lib/utils/variants.ts` | derivative URL convention (see below) |
| `src/lib/components/GalleryGrid.svelte` | offset 12-column grid, per-set sections |
| `src/lib/components/PhotoPlate.svelte` | one photograph: placeholder, image, caption |
| `src/lib/components/Lightbox.svelte` | viewer with EXIF panel, keyboard + swipe |
| `src/lib/components/SplitText.svelte` | sliced-text assembly used for the wordmark |
| `src/lib/components/Atmosphere.svelte` | grain + light leak overlays |
| `metadata-api/` | the Cloudflare Worker that serves the metadata API |
| `scripts/` | Go tooling that uploads images and their metadata to R2 |

## Resized images

The bucket holds full-resolution originals only; the gallery never loads them.
Each original has three WebP derivatives under a `d/` prefix — `d/w800` for grid
cells, `d/w2200` for the lightbox, `d/lqip` for the blur-up placeholder — and the
Worker filters that prefix out of its listing. `src/lib/utils/variants.ts` derives
the URLs from the original key; `metadata-api/README.md` documents the table.

Generating and uploading them is out of the app's hands: whatever uploads an
original must also upload those three keys.

## Theme

Dark by default. The palette, type scale and grid tokens are CSS custom properties
in `src/app.css`, switched by `data-theme` on `<html>`; a small inline script in
`src/app.html` resolves the stored preference before first paint.

## Deploy

`.github/workflows/deploy.yml` builds the app on push to `main` and publishes
`build/` to GitHub Pages (custom domain `art.fahadfaruqi.com`). The Worker is
deployed separately:

```sh
npm run worker:deploy
```
