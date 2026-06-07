document.addEventListener('click', function (e) {
  const card = e.target.closest('.pull-request');
  if (!card) return;

  const hasBody = card.dataset.hasBody === 'true';
  if (!hasBody) return;

  const prBody = card.querySelector('.pr-body');
  const icon = card.querySelector('.button-toggle i');
  const isChevron = e.target.closest('.button-toggle') !== null;

  if (isChevron) {
    if (icon.classList.contains('fa-chevron-down')) {
      icon.classList.replace('fa-chevron-down', 'fa-chevron-up');
      prBody.classList.add('pr-body--open');
      card.classList.add('is-open');
    } else {
      icon.classList.replace('fa-chevron-up', 'fa-chevron-down');
      prBody.classList.remove('pr-body--open');
      card.classList.remove('is-open');
    }
  } else if (icon.classList.contains('fa-chevron-down')) {
    icon.classList.replace('fa-chevron-down', 'fa-chevron-up');
    prBody.classList.add('pr-body--open');
    card.classList.add('is-open');
  }
});
