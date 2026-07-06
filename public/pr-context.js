let renderedDiffs = [];
let currentAbort = null;

function getStateLabel(state) {
  if (state === "Merged") return "Merged";
  if (state === "Draft") return "Draft";
  if (state === "Closed") return "Closed";
  return "Open";
}

function getStateBadge(state) {
  if (state === "Merged") return "badge-merged";
  if (state === "Draft") return "badge-neutral";
  if (state === "Closed") return "badge-danger";
  return "badge-success";
}

function escapeHTML(str) {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function getOwnerRepo() {
  var parts = window.location.pathname.split("/").filter(Boolean);
  return { owner: parts[0] || "", repo: parts[1] || "" };
}

function buildPRHeaderFragment(card) {
  var number = card.dataset.prNumber;
  var title = card.dataset.prTitle;
  var author = card.dataset.prAuthor;
  var state = card.dataset.prState;

  var fragment = document.createDocumentFragment();

  var header = document.createElement("div");
  header.className = "pr-detail-header";
  header.innerHTML = '<div class="pr-detail-title">' + escapeHTML(title) + '</div><div class="pr-detail-number">#' + escapeHTML(number) + "</div>";
  fragment.appendChild(header);

  var meta = document.createElement("div");
  meta.className = "pr-detail-meta";
  meta.innerHTML = '<span class="' + getStateBadge(state) + '" style="padding:2px 8px;border-radius:12px;font-size:12px;font-weight:500">' + getStateLabel(state) + "</span>" + '<span>' + escapeHTML(author) + "</span>";
  fragment.appendChild(meta);

  var bodyEl = card.querySelector(".pr-body > .pr-body-inner, .pr-body");
  if (bodyEl && bodyEl.innerHTML.trim()) {
    var bodySection = document.createElement("div");
    bodySection.className = "pr-detail-body";
    bodySection.innerHTML = '<div class="section-header" style="margin-bottom:12px"><span class="section-title">Description</span></div><div class="pr-detail-body-content">' + bodyEl.innerHTML + "</div>";
    fragment.appendChild(bodySection);
  }

  return fragment;
}

function openPRInPane(card) {
  if (currentAbort) currentAbort.abort();
  var ac = new AbortController();
  currentAbort = ac;
  var signal = ac.signal;

  renderedDiffs.forEach(function (d) {
    try { d.destroy(); } catch (e) {}
  });
  renderedDiffs = [];

  var container = document.getElementById("right-pane-content");
  var rightPane = document.getElementById("right-pane");
  if (!container || !rightPane) return;

  rightPane.classList.add("open");

  container.innerHTML = "";
  container.appendChild(buildPRHeaderFragment(card));

  var number = card.dataset.prNumber;
  var ownerRepo = getOwnerRepo();

  var diffsSection = document.createElement("div");
  diffsSection.className = "pr-detail-diffs";
  diffsSection.innerHTML = '<div class="section-header" style="margin-bottom:12px"><span class="section-title">Changes</span></div><div class="pr-diff-loading">Loading diffs\u2026</div>';
  container.appendChild(diffsSection);

  import("./pr-diffs.js").then(function (mod) {
    mod.loadPRDiffs(diffsSection, ownerRepo.owner, ownerRepo.repo, number, signal).then(function (diffs) {
      diffs.forEach(function (d) { renderedDiffs.push(d); });
    });
  }).catch(function () {
    var loadingEl = diffsSection.querySelector(".pr-diff-loading");
    if (loadingEl) loadingEl.remove();
    diffsSection.innerHTML += '<div class="pr-diff-error">Failed to load diff viewer</div>';
  });

  if (window.innerWidth <= 768) {
    var tab = document.querySelector('[data-view="right"]');
    if (tab) tab.click();
  }
}

document.addEventListener("click", function (e) {
  var card = e.target.closest(".pull-request");
  if (!card) {
    if (e.target.closest("body:not(.repo-page)")) return;
    return;
  }
  if (e.target.closest("a") || e.target.closest(".pr-body")) return;

  document.querySelectorAll(".pull-request").forEach(function (c) {
    c.classList.remove("selected");
  });
  card.classList.add("selected");

  openPRInPane(card);
});
