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

`-dir` and `-pattern` are joined the way a path is, so `-d paintings -p '*.jpg'`
selects the keys under `paintings/`. A glob is matched with Go's `filepath.Match`
rules, so `*` does not cross a `/`: `-p '*'` selects only the root-level keys,
while `-d d/w800 -p '*.webp'` selects one derivative variant. Alternatively pass
bucket keys as positional arguments.

## Commands

| Command | What it does |
| --- | --- |
| `create [file...]` | Extracts EXIF, prompts for the required metadata, uploads |
| `read [key...]` | Prints the CDN URL of each matching object |
| `update [key...]` | Edits metadata, and renames the object when `-name` is given |
| `delete [key...]` | Deletes the matching objects, after a confirmation prompt |

### create

```sh
./manage-images create photo.jpg
./manage-images create -d ~/Pictures/raws -p '*.jpg'
```

Prompts for `title`, `altText`, `description`, `set`, and `number`, reusing the
first image's set as the default for the rest of the batch. The object key is the
local file name, so the upload lands at the bucket root.

### read

```sh
./manage-images read                      # every object
./manage-images read -p 'sam-*.png'
./manage-images read -d d/lqip -p '*.webp'
```

One URL per line, sorted, so it pipes into `xargs`, `curl`, or a browser.

### update

```sh
./manage-images update sam-1.png --title 'Morning light' --number 3
./manage-images update -p 'sam-*.png' --set 'Sam'
./manage-images update sam-1.png --name sam-01        # -> sam-01.png
```

Only the flags you pass are changed; the rest of the object's metadata is
preserved. `-name` takes a new file name, keeps the original extension when the
new name has none, and stays inside the object's own prefix. Renaming copies the
object and deletes the old key, which leaves the `d/<variant>/…` derivatives this
name is derived from behind — re-run `node scripts/derivatives.mjs` afterwards.

### delete

```sh
./manage-images delete sam-1.png
./manage-images delete -p 'sam-*.png' -y
```

Lists what it is about to remove and waits for `y` unless `-y/--yes` is passed.
Deleting an original leaves its derivatives in place.
