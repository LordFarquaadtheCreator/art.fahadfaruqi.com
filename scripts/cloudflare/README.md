# Manage images

A Go CLI for the R2 bucket behind `https://assets.fahadfaruqi.com`, covering the
full lifecycle of an image: create, read, update, delete.

```sh
cd scripts
go build -o manage-images ./cmd/manage-images
```

## Config

`config.yaml` (gitignored) holds the R2 credentials:

```yaml
r2:
  account_id: ...
  bucket: assets
  access_key_id: ...
  secret_access_key: ...
  s3_api_endpoint: https://<account_id>.r2.cloudflarestorage.com
  cdn_base: https://assets.fahadfaruqi.com
```

`cdn_base` is the public origin the bucket is served from, used by `read` to build
URLs. `CDN_BASE` in the environment overrides it.

## Selecting which objects a command works on

Every command takes the same two flags:

| Flag | Meaning |
| --- | --- |
| `-d, --dir` | Local directory for `create`; key prefix in the bucket for the rest |
| `-p, --pattern` | Glob, matched against local file names or bucket keys |

`-dir` and `-pattern` are joined the way a path is, so `-d art/master -p '*.webp'`
selects every master. A glob is matched with Go's `filepath.Match` rules, so `*`
does not cross a `/`: `-p '*'` selects only the keys directly under the prefix,
while `-d art/compressed -p '*.avif'` selects the compressed tier. Alternatively
pass bucket keys as positional arguments.

## Commands

| Command | What it does |
| --- | --- |
| `create [file...]` | Extracts EXIF, prompts for the required metadata, writes the master and its compressed sibling; `--inherit <ext>` takes both from the same-named object instead |
| `read [key...]` | Prints the CDN URL of each matching object |
| `update [key...]` | Edits metadata, and renames the object when `-name` is given |
| `delete [key...]` | Deletes the matching objects, after a confirmation prompt |

### create

```sh
./manage-images create photo.jpg
./manage-images create -d ~/Pictures/raws -p '*.jpg'
```

Prompts for `title`, `altText`, `description`, `set`, and `number`, reusing the
first image's set as the default for the rest of the batch.

Each file is published as two objects: the master under `art/master/<name>`, and a
compressed sibling under `art/compressed/<stem>.avif` — same stem, AVIF, 1600 px
wide with the height following the master's ratio, encoded by
`ffmpeg -c:v libsvtav1`. The gallery renders the compressed copy and fetches the
master on hover, so both are written by the one command. If `ffmpeg` or `ffprobe`
is missing, nothing is uploaded: a master without its sibling is a photograph the
gallery would have to render at full size. Both objects carry the same curated
metadata and EXIF, plus the `width` and `height` of the file itself.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-master-prefix` | `art/master` | Key prefix the master is written under |
| `-compressed-prefix` | `art/compressed` | Key prefix the compressed sibling is written under |
| `-compressed-width` | `1600` | Width of the sibling; the height follows the master's ratio |
| `-crf` | `22` | AVIF crf for the sibling — lower is better and larger |
| `-preset` | `6` | SVT-AV1 preset for the sibling — slower presets spend quality, not bytes |

Cache-Control is set at upload: `public, max-age=86400` on the master and
`public, max-age=604800` on the compressed copy. Neither is `immutable`, so a
re-upload under the same key heals within a day or a week instead of sticking in
browsers for a year.

`--inherit <ext>` uploads a re-encoded file with the metadata of the bucket object it
replaces — the same name carrying that extension, e.g. `--inherit .png` for a WebP made
from a PNG, read from `art/master/` — and skips the prompts entirely. It also carries
the original's upload time into custom metadata, because R2's own timestamp resets on
every upload and the gallery reads that value as the set's date.

`EXIF.Extract` reads by extension and has no `.webp` case, so a WebP uploaded without
`--inherit` arrives with an empty EXIF map (and a warning). Anything re-encoded from an
object that already carries metadata should go through `--inherit`.

```sh
./manage-images create -d ~/exports/webp -p '*.webp' --inherit .png
```

### read

```sh
./manage-images read                                      # every object
./manage-images read -d art/master -p 'sam-*.webp'
./manage-images read -d art/compressed -p '*.avif'
```

One URL per line, sorted, so it pipes into `xargs`, `curl`, or a browser.

### update

```sh
./manage-images update -d art/master sam-1.webp --title 'Morning light' --number 3
./manage-images update -d art/master -p 'sam-*.webp' --set 'Sam'
./manage-images update -d art/master sam-1.webp --name sam-01     # -> art/master/sam-01.webp
```

Only the flags you pass are changed; the rest of the object's metadata is
preserved. `-name` takes a new file name, keeps the original extension when the
new name has none, and stays inside the object's own prefix. Renaming copies the
object and deletes the old key, which leaves the compressed sibling of the old stem
in `art/compressed/` behind, where the Worker now pairs it with nothing — rename or
re-publish that side too.

### delete

```sh
./manage-images delete -d art/master sam-1.webp
./manage-images delete -d art/master -p 'sam-*.webp' -y
```

Lists what it is about to remove and waits for `y` unless `-y/--yes` is passed.
Deleting a master leaves its compressed sibling in place, and the other way round.
