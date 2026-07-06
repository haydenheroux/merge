function loadImages(container) {
  container.querySelectorAll('img[data-src]').forEach(function (img) {
    img.src = img.dataset.src;
    img.removeAttribute('data-src');
  });
}

document.addEventListener('click', function (e) {
  const card = e.target.closest('.pull-request');
  if (!card) return;

  const hasBody = card.dataset.hasBody === 'true';
  if (!hasBody) return;

  const prBody = card.querySelector('.pr-body');
  if (!prBody) return;

  const isToggle = e.target.closest('.pr-body') !== null;

  if (!isToggle) {
    prBody.classList.toggle('is-open');
    if (prBody.classList.contains('is-open')) {
      loadImages(prBody);
    }
  }
});
