# Project Instructions

Small Go forum image uploader. Uses Go `net/http`, browser HTMX, `github.com/joho/godotenv` for local env.

## Commands

- Dev server: `make run`.
- Tests: `make test`.
- Build binary: `make build`.
- Build Docker image: `make docker-build`.
- Format Go: `make fmt`.
- Tidy deps: `make tidy`.
- Remove build output: `make clean`.

## Runtime Configuration

- Secrets live in `.env`.
- Never commit `.env`.
- Document env names in `.env.example`.
- `IMGBB_API_KEY` server-side only. Never expose in HTML, JavaScript, CSS, logs, client responses.
- `ADDR` controls listen addr. Default `:8080`.

## Security Rules

- Upload handlers = untrusted input boundary.
- Keep API keys + upstream service errors out of browser responses.
- Validate uploaded image bytes with `http.DetectContentType`.
- Keep upload size limits + rate limiting unless product req changes them.
- Preserve security headers when changing middleware/routing.
- Do not log secrets, req bodies, image contents, full upstream responses.

## Architecture Notes

- `main.go`: app setup, routes, index handler, env helper.
- `upload.go`: upload handler, image validation, filename cleanup, upload error fragments.
- `providers.go`: Catbox/imgbb upload providers, fallback order, uploaded URL validation.
- `limiter.go`: per-IP upload throttling.
- `middleware.go`: security headers.
- `templates.go`: HTML fragments/templates.
- `Dockerfile`: multi-stage container build. Runtime image includes binary + `static/`, runs as non-root `app`.
- `static/app.js`: HTMX upload UI behavior, clipboard copy.
- `static/styles.css`: forum-like visual style.
- Upload order: Catbox first, imgbb second.
- Generated forum text must be `[img src=https://example.com/image.png]`. No double quotes around URL.

## UI Rules

- Match `https://forum.sodika.dk/index.php#bottom` colors: light gray page, white panels, gray borders, blue links, neutral gray buttons, plain black text.
- Keep UI usable mobile + desktop.
- Do not restore Sodipedia article recommendation flow unless explicitly requested.

## Editing Rules

- Keep changes small, aligned with existing files.
- Prefer stdlib unless existing dep solves problem.
- Use `gofmt` for Go files.
- Run `make test` after behavior changes.
- Run `make build` after startup/dependency changes.
