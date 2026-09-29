# [art.fahadfaruqi.com](art.fahadfaruqi.com)

Photographs by [Fahad Faruqi](https://fahadfaruqi.com) — a static gallery built with
SvelteKit, served from GitHub Pages, subdomain managed by Cloudflare. The photographs 
themselves live in Cloudflare R2, served straight from the bucket for the images and 
through a small Worker for the listing.

## Why it is built this way

By storing metadata in an image - we can use our CDN's free tier as our database. 

```
R2 bucket "assets"
 ├── art/master/<name>.webp          masters (fetched on hover, drawn when opened)
 └── art/compressed/<stem>.avif      the file every view renders
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

`.env` (gitignored) may override `VITE_METADATA_API` and `VITE_CDN_BASE` for builds;
when unset, the production URLs are used, which is what CI does. `bun run dev` ignores
both and reads through the dev server, whose proxy fetches everything uncached — a
photograph uploaded a minute ago is in the next reload, not in ten minutes.

## Images and metadata

Every photograph is two objects: a master and its compressed sibling, the same stem
apart, written together by `scripts/cloudflare`'s `create`:

| Key | Format | Used for |
| --- | --- | --- |
| `art/master/<name>.webp` | WebP q85, 6016 px | the archival file; fetched on hover, rendered in the viewer |
| `art/compressed/<stem>.avif` | AVIF, 1600 px | every view: the grid cells and the viewer's underlay |

Captions come from custom metadata on both — `set`, `number`, `title`, `alttext`,
`description` — alongside the EXIF the camera wrote. The Worker lists the `art/` prefix,
pairs the two objects by stem, and returns both URLs plus the dimensions the grid needs
before an image arrives. `metadata-api/README.md` documents the API, `scripts/README.md`
the CLI that writes both objects, and `AGENTS.md` the details of working in this repo.

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
