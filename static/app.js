(function () {
  const form = document.getElementById("upload-form");
  const finalPanel = document.getElementById("final-panel");
  const resultHolder = document.getElementById("result-holder");

  function showUploadingState() {
    finalPanel.hidden = true;
    resultHolder.innerHTML = '<div class="result uploading"><div class="spinner" aria-hidden="true"></div><strong>Feltoltes folyamatban...</strong></div>';
    finalPanel.hidden = false;
  }

  document.body.addEventListener("htmx:beforeRequest", function (event) {
    if (event.target !== form) return;
    showUploadingState();
  });

  document.body.addEventListener("htmx:afterSwap", function (event) {
    if (event.target !== resultHolder) return;
    finalPanel.hidden = false;
    finalPanel.scrollIntoView({ behavior: "smooth", block: "start" });
  });

  document.body.addEventListener("htmx:sendError", function (event) {
    if (event.target !== form) return;
    resultHolder.innerHTML = '<div class="result error"><strong>Hiba</strong><p>A feltoltesi keres nem ment at. Probald ujra.</p></div>';
    finalPanel.hidden = false;
  });

  document.addEventListener("click", function (event) {
    const button = event.target.closest("[data-copy-target]");
    if (!button) return;
    const target = document.getElementById(button.dataset.copyTarget);
    if (!target) return;

    navigator.clipboard.writeText(target.value).then(
      function () {
        const original = button.textContent;
        button.textContent = "Masolva";
        setTimeout(function () {
          button.textContent = original;
        }, 1400);
      },
      function () {
        target.select();
        document.execCommand("copy");
      }
    );
  });
})();
