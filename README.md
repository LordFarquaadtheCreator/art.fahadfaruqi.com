# [art.fahadfaruqi.com](art.fahadfaruqi.com)

Photographs by [Fahad Faruqi](https://fahadfaruqi.com) — a static gallery built with
SvelteKit, served from GitHub Pages, subdomain managed by Cloudflare. The photographs 
themselves live in Cloudflare R2, served straight from the bucket for the images and 
through a small Worker for the listing.

## Why it is built this way

By storing metadata in an image - we can use our CDN's free tier as our database. 

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
(`.github/workflows/deploy.yml` → GitHub Pages). 
The Worker/api deploys separately:

```sh
bun run worker:deploy
```

## Design credit

The layout follows the Stefan Vitasović portfolio case study published on Codrops
(2025). The performance is still the cheap one: every photograph in the gallery is a
plain `<img>` that a browser without WebGL renders on its own, and the layer above it is
an effect rather than the delivery mechanism. The print treatment and the pointer-tracked
light are this site's own, built on that reference rather than copied from it.
