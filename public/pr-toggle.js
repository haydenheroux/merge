document.addEventListener('click', function (e) {
  const button = e.target.closest('.button-toggle');
  if (!button) return;

  const prBody = button.closest('.pull-request').querySelector('.pr-body');
  const hasBody = prBody && prBody.innerHTML.trim().length > 0;
  const icon = button.querySelector('i');

  if (icon.classList.contains('fa-chevron-down')) {
    icon.classList.replace('fa-chevron-down', 'fa-chevron-up');
    if (hasBody) prBody.classList.add('pr-body--open');
  } else {
    icon.classList.replace('fa-chevron-up', 'fa-chevron-down');
    if (hasBody) prBody.classList.remove('pr-body--open');
  }
});
