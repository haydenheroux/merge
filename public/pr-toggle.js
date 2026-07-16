var currentPrNumber = null;

function loadImages(container) {
  container.querySelectorAll('img[data-src]').forEach(function (img) {
    img.src = img.dataset.src;
    img.removeAttribute('data-src');
  });
}

function openPane(prNumber) {
  var pane = document.getElementById('right-pane');
  var content = document.getElementById('right-pane-content');
  var main = document.querySelector('main');

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

  content.innerHTML = '<div class="pane-right-body skeleton"><div class="pane-right-header"><div class="skeleton-line" style="width: 60%; height: $text-lg;"></div></div><div class="pane-right-meta"><div class="skeleton-line" style="width: 40%;"></div><div class="skeleton-line" style="width: 30%;"></div><div class="skeleton-line" style="width: 35%;"></div></div><div class="pr-body-inner"><div class="skeleton-line" style="width: 90%;"></div><div class="skeleton-line" style="width: 75%;"></div><div class="skeleton-line" style="width: 85%;"></div><div class="skeleton-line" style="width: 60%;"></div></div></div>';

  if (owner && repo) {
    fetch('/' + owner + '/' + repo + '?part=pr-detail&number=' + prNumber)
      .then(function(response) {
        return response.text();
      })
      .then(function(html) {
        content.innerHTML = html;
        loadImages(content);
        var closeBtn = content.querySelector('.button-toggle');
        if (closeBtn) {
          closeBtn.addEventListener('click', function(ev) {
            ev.stopPropagation();
            closePane();
          });
        }
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
  if (main) main.classList.remove('right-pane-open');
}

document.addEventListener('click', function (e) {
  var card = e.target.closest('.pull-request');
  if (!card) {
    if (e.target.closest('.pane-right')) return;
    closePane();
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
    closePane();
  }
});
