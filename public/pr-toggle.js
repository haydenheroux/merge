var currentPrNumber = null;
var filterPanelOpen = false;
var currentDiffInstances = [];
var currentDiffAbort = null;

function loadImages(container) {
  container.querySelectorAll('img[data-src]').forEach(function (img) {
    var src = img.dataset.src;
    
    // Create a new image to preload
    var preloadImg = new Image();
    preloadImg.onload = function() {
      img.src = src;
      img.removeAttribute('data-src');
      img.classList.add('loaded');
    };
    preloadImg.src = src;
  });
}

// --- PR Detail Pane ---

function openPane(prNumber) {
  var pane = document.getElementById('right-pane');
  var content = document.getElementById('right-pane-content');
  var main = document.querySelector('main');

  closeFilterPanel();
  cleanupCurrentDiffs();

  var prevSelected = document.querySelector('.pull-request.selected');
  if (prevSelected) {
    prevSelected.classList.remove('selected');
  }

  var card = document.querySelector('.pull-request[data-pr-number="' + prNumber + '"]');
  if (card) {
    card.classList.add('selected');
  }

  var owner = card ? card.closest('[data-owner]')?.dataset.owner : null;
  var repo = card ? card.closest('[data-repo]')?.dataset.repo : null;

  if (!owner || !repo) {
    var pathParts = window.location.pathname.split('/').filter(Boolean);
    if (pathParts.length >= 2) {
      owner = pathParts[0];
      repo = pathParts[1];
    }
  }

  content.innerHTML = '<div class="pane-right-header"><div class="skeleton-line" style="width: 60%; height: $text-lg;"></div></div><div class="pane-right-body skeleton scroll"><div class="pane-right-meta"><div class="skeleton-line" style="width: 40%;"></div><div class="skeleton-line" style="width: 30%;"></div><div class="skeleton-line" style="width: 35%;"></div></div><div class="skeleton-line" style="width: 90%;"></div><div class="skeleton-line" style="width: 75%;"></div><div class="skeleton-line" style="width: 85%;"></div><div class="skeleton-line" style="width: 60%;"></div></div>';

  if (owner && repo) {
    var abortCtrl = new AbortController();
    currentDiffAbort = abortCtrl;

    fetch('/' + owner + '/' + repo + '?part=pr-detail&number=' + prNumber)
      .then(function(response) {
        return response.text();
      })
      .then(function(html) {
        if (abortCtrl.signal.aborted) return;
        content.innerHTML = html;
        loadImages(content);
        var closeBtn = content.querySelector('.button-toggle');
        if (closeBtn) {
          closeBtn.addEventListener('click', function(ev) {
            ev.stopPropagation();
            closePane();
          });
        }

        loadDiffsInPane(content, owner, repo, prNumber, abortCtrl.signal);
      });
  }

  pane.classList.add('open');
  currentPrNumber = prNumber;
  if (main) main.classList.add('right-pane-open');
}

function closePane() {
  var pane = document.getElementById('right-pane');
  var content = document.getElementById('right-pane-content');
  var main = document.querySelector('main');
  var selected = document.querySelector('.pull-request.selected');

  if (selected) {
    selected.classList.remove('selected');
  }

  content.innerHTML = '';
  pane.classList.remove('open');
  currentPrNumber = null;
  cleanupCurrentDiffs();
  if (main) main.classList.remove('right-pane-open');
}

function cleanupCurrentDiffs() {
  if (currentDiffAbort) {
    currentDiffAbort.abort();
    currentDiffAbort = null;
  }
  if (currentDiffInstances.length) {
    import('./pr-diffs.js').then(function(mod) {
      mod.cleanupDiffs(currentDiffInstances);
    }).catch(function() {});
    currentDiffInstances = [];
  }
}

function loadDiffsInPane(content, owner, repo, prNumber, signal) {
  import('./pr-diffs.js').then(function(mod) {
    var target = content.querySelector('.pane-right-body') || content;

    var skeleton = document.createElement('div');
    skeleton.className = 'pr-diff-skeleton';
    skeleton.innerHTML = '<div class="skeleton-line" style="width: 100%; height: 12px; margin-bottom: 8px;"></div><div class="skeleton-line" style="width: 90%; height: 12px; margin-bottom: 8px;"></div><div class="skeleton-line" style="width: 95%; height: 12px; margin-bottom: 8px;"></div><div class="skeleton-line" style="width: 85%; height: 12px; margin-bottom: 8px;"></div><div class="skeleton-line" style="width: 88%; height: 12px;"></div>';
    target.appendChild(skeleton);

    return mod.loadPRDiffs(target, owner, repo, prNumber, signal).then(function(instances) {
      if (!signal.aborted) {
        currentDiffInstances = instances;
        skeleton.remove();
      }
    });
  }).catch(function(err) {
    console.error('[diffs]', err);
    if (!signal.aborted) {
      var skeleton = content.querySelector('.pr-diff-skeleton');
      if (skeleton) {
        skeleton.innerHTML = '<div style="padding: 1rem; color: var(--clr-text-muted); font-size: 0.875rem; text-align: center;">Failed to load diffs</div>';
      }
    }
  });
}

// --- Filter Panel ---

function moveFiltersToPanel() {
  var panel = document.getElementById('filter-panel');
  if (!panel) return;
  var panelBody = panel.querySelector('.filter-panel-body');
  if (!panelBody) return;

  var scopesSection = document.getElementById('scopes-section');
  var contribsSection = document.getElementById('contributors-section');

  if (scopesSection && scopesSection.parentElement !== panelBody) {
    panelBody.appendChild(scopesSection);
  }
  if (contribsSection && contribsSection.parentElement !== panelBody) {
    panelBody.appendChild(contribsSection);
  }
}

function openFilterPanel() {
  var panel = document.getElementById('filter-panel');
  if (!panel) return;

  closePane();
  moveFiltersToPanel();

  panel.classList.add('open');
  filterPanelOpen = true;
}

function closeFilterPanel() {
  var panel = document.getElementById('filter-panel');
  if (!panel) return;
  panel.classList.remove('open');
  filterPanelOpen = false;
}

// --- Event Listeners (delegated) ---

document.addEventListener('click', function (e) {
  // Filter toggle button
  if (e.target.closest('#filter-toggle')) {
    e.stopPropagation();
    if (filterPanelOpen) {
      closeFilterPanel();
    } else {
      openFilterPanel();
    }
    return;
  }

  // Filter close button
  if (e.target.closest('#filter-close')) {
    e.stopPropagation();
    closeFilterPanel();
    return;
  }

  // PR card clicks
  var card = e.target.closest('.pull-request');
  if (!card) {
    if (e.target.closest('.pane-right')) return;
    if (e.target.closest('.filter-panel')) return;
    closePane();
    closeFilterPanel();
    return;
  }

  if (e.target.closest('a')) return;

  var hasBody = card.dataset.hasBody === 'true';
  if (!hasBody) return;

  var prNumber = parseInt(card.dataset.prNumber, 10);
  if (isNaN(prNumber)) return;

  if (prNumber === currentPrNumber) {
    closePane();
    return;
  }

  openPane(prNumber);
});

document.addEventListener('keydown', function (e) {
  if (e.key === 'Escape') {
    if (filterPanelOpen) {
      closeFilterPanel();
    } else {
      closePane();
    }
  }
});

document.addEventListener('htmx:afterSwap', function (e) {
  // Re-open filter panel if it was open before the swap
  if (filterPanelOpen) {
    openFilterPanel();
  }

  // Re-select the current PR card
  if (currentPrNumber !== null) {
    var card = document.querySelector('.pull-request[data-pr-number="' + currentPrNumber + '"]');
    if (card) {
      card.classList.add('selected');
    }
  }

  // Load deferred images after HTMX swap
  loadImages(e.detail.target || document);
});

// Load deferred images on initial page load
document.addEventListener('DOMContentLoaded', function() {
  loadImages(document);
});
