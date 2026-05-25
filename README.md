# sodiforum-img

Small Go + HTMX web app for uploading images and generating SoDIfórum-ready image markup.

The app uploads to Catbox first. If Catbox fails, it falls back to imgbb. After upload, it generates:

```text
[img src=https://example.com/image.png]
```

## Features

- Image upload from mobile or desktop.
- Catbox-first upload flow with imgbb fallback.
- Copy button for generated forum text.
- Forum-like colors and layout matching `https://forum.sodika.dk/index.php#bottom`.
- Basic upload rate limiting.
- Byte-based image validation with `http.DetectContentType`.
- Security headers for browser responses.
- Docker image support.

## Requirements

- Go 1.22+
- `make`
- Docker, only if building/running container image

## Configuration

Create `.env` from `.env.example`:

```sh
cp .env.example .env
```

Set:

```env
IMGBB_API_KEY=your-imgbb-api-key
```

Optional:

```env
ADDR=:8080
```

Do not commit `.env`. It contains secrets.

## Run Locally

```sh
make run
```

Open:

```text
http://localhost:8080
```

## Development Commands

```sh
make fmt
make test
make build
make tidy
make clean
```

## Docker

Build image:

```sh
make docker-build
```

Run container:

```sh
docker run --rm -p 8080:8080 --env-file .env sodiforum-img:latest
```

Open:

```text
http://localhost:8080
```

## Project Structure

- `main.go`: app setup, routes, index handler, env helper.
- `upload.go`: upload handler, image validation, filename cleanup, upload error fragments.
- `providers.go`: Catbox/imgbb upload providers, fallback order, uploaded URL validation.
- `limiter.go`: per-IP upload throttling.
- `middleware.go`: security headers.
- `templates.go`: HTML templates/fragments.
- `static/app.js`: HTMX upload UI behavior and clipboard copy.
- `static/styles.css`: forum-like visual style.
- `Dockerfile`: multi-stage container build running as non-root user.

## Security Notes

- `IMGBB_API_KEY` is server-side only.
- API keys and upstream service errors are not returned to the browser.
- Uploads are limited to 12 MB.
- Allowed detected image types: JPEG, PNG, GIF, WebP.
- Uploaded provider URLs must be HTTPS and must not contain quotes, brackets, spaces, or control whitespace.
