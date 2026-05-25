package main

const resultHTML = `
<div id="upload-result" class="result" data-ready="true">
  <div class="result-meta">Feltoltve: %s</div>
  <label for="bbcode">Forum kod</label>
  <div class="copy-row">
    <textarea id="bbcode" readonly rows="2">%s</textarea>
    <button class="button secondary" type="button" data-copy-target="bbcode">Masol</button>
  </div>
  <a class="plain-link" href="%s" target="_blank" rel="noreferrer">Kep megnyitasa</a>
</div>`

const indexHTML = `<!doctype html>
<html lang="hu">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SoDIfórum képfeltöltő</title>
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
  <header class="topbar">
    <div class="shell nav">
      <a class="brand" href="/">SoDIfórum</a>
      <nav aria-label="Forum menu">
        <a href="https://forum.sodika.dk/">Home</a>
      </nav>
    </div>
  </header>

  <main class="shell layout">
    <section class="composer" aria-labelledby="title">
      <div class="composer-head">
        <p class="eyebrow">Kepfeltolto</p>
        <h1 id="title">Forumkep kod egy feltoltesbol</h1>
      </div>

      <form id="upload-form" class="upload-form" hx-post="/upload" hx-target="#result-holder" hx-swap="innerHTML" hx-encoding="multipart/form-data">
        <label class="dropzone" for="image">
          <input id="image" name="image" type="file" accept="image/*" required>
          <span class="drop-title">Valassz kepet</span>
          <span class="drop-note">JPG, PNG, GIF vagy WebP, maximum 12 MB</span>
        </label>
        <button class="button primary" type="submit">Feltoltes</button>
      </form>
    </section>

    <section id="final-panel" class="final-panel" hidden>
      <div id="result-holder"></div>
    </section>
  </main>

  <script src="/static/app.js"></script>
</body>
</html>`
