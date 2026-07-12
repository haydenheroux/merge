function loadImages(container) {
  container.querySelectorAll('img[data-src]').forEach(function (img) {
    img.src = img.dataset.src;
    img.removeAttribute('data-src');
  });
}

function closePane() {
  var pane = document.getElementById('right-pane');
  var content = document.getElementById('right-pane-content');
  var main = document.querySelector('main');
  var selected = document.querySelector('.pull-request.selected');

  if (selected) {
    var body = selected.querySelector('.pr-body');
    if (body && content.firstChild) {
      body.appendChild(content.firstChild);
    }
    selected.classList.remove('selected');
  }

  content.innerHTML = '';
  pane.classList.remove('open');
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

  var main = document.querySelector('main');
  var pane = document.getElementById('right-pane');
  var content = document.getElementById('right-pane-content');

  if (card.classList.contains('selected')) {
    closePane();
    return;
  }

  var prevSelected = document.querySelector('.pull-request.selected');
  if (prevSelected) {
    var prevBody = prevSelected.querySelector('.pr-body');
    if (prevBody && content.firstChild) {
      prevBody.appendChild(content.firstChild);
    }
    prevSelected.classList.remove('selected');
  }

  var prBody = card.querySelector('.pr-body');
  if (prBody) {
    var inner = prBody.querySelector('.pr-body-inner');
    if (inner) {
      content.innerHTML = '';
      var wrapper = document.createElement('div');
      wrapper.className = 'pane-right-body';
      var header = document.createElement('div');
      header.className = 'pane-right-header';
      var titleText = document.createElement('span');
      titleText.className = 'pane-right-title';
      titleText.textContent = card.querySelector('.title') ? card.querySelector('.title').textContent.trim() : 'Pull Request';
      var closeBtn = document.createElement('button');
      closeBtn.className = 'button-toggle';
      closeBtn.setAttribute('aria-label', 'Close');
      closeBtn.innerHTML = '<i class="fa-solid fa-xmark" aria-hidden="true"></i>';
      closeBtn.addEventListener('click', function (ev) {
        ev.stopPropagation();
        closePane();
      });
      header.appendChild(titleText);
      header.appendChild(closeBtn);
      wrapper.appendChild(header);
      var descRow = card.querySelector('.description-row');
      if (descRow) {
        var meta = document.createElement('div');
        meta.className = 'pane-right-meta';
        var spans = descRow.querySelectorAll(':scope > span');
        var icons = ['fa-code-pull-request', 'fa-clock', 'fa-user', 'fa-tag'];
        var labels = ['Status', 'Age', 'Author', 'Scope'];
        for (var i = 0; i < spans.length; i++) {
          var row = document.createElement('div');
          row.className = 'meta-row';
          var label = document.createElement('span');
          label.className = 'meta-label';
          var icon = document.createElement('i');
          icon.className = 'fa-solid ' + (icons[i] || 'fa-tag');
          label.appendChild(icon);
          label.appendChild(document.createTextNode(' ' + (labels[i] || 'Info')));
          var value = document.createElement('span');
          value.className = 'meta-value';
          value.innerHTML = spans[i].innerHTML;
          row.appendChild(label);
          row.appendChild(value);
          meta.appendChild(row);
        }
        wrapper.appendChild(meta);
      }
      var bodyClone = inner.cloneNode(true);
      wrapper.appendChild(bodyClone);
      content.appendChild(wrapper);
      loadImages(wrapper);
    }
  }

  card.classList.add('selected');
  pane.classList.add('open');
  if (main) main.classList.add('right-pane-open');
});

document.addEventListener('keydown', function (e) {
  if (e.key === 'Escape') {
    closePane();
  }
});
