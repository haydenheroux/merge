// Diff library loading
let diffsLib = null;
let diffsLibPromise = null;

async function ensureDiffsLib() {
  if (diffsLib) return diffsLib;
  // If already loading, wait for that promise
  if (diffsLibPromise) return diffsLibPromise;
  // Start loading and cache the promise
  diffsLibPromise = import('https://esm.sh/@pierre/diffs@1.2.12').then(lib => {
    diffsLib = lib;
    return lib;
  });
  return diffsLibPromise;
}

function getDiffTheme() {
  return document.documentElement.classList.contains('dark') ? 'pierre-dark' : 'pierre-light';
}

async function loadPRDiffs(container, owner, repo, prNumber, signal) {
  const lib = await ensureDiffsLib();

  const resp = await fetch(`/${owner}/${repo}/pull/${prNumber}/diff`, { signal });
  if (!resp.ok) throw new Error('Failed to fetch diff: ' + resp.status);

  const diffText = await resp.text();
  const patches = lib.parsePatchFiles(diffText);

  const instances = [];

  for (const patch of patches) {
    for (const fileDiffMeta of patch.files) {
      const fileDiff = new lib.FileDiff({
        theme: getDiffTheme(),
        diffStyle: 'unified',
        overflow: 'scroll',
        disableLineNumbers: false,
        diffIndicators: 'bars',
      });

      fileDiff.render({ fileDiff: fileDiffMeta, containerWrapper: container });
      instances.push(fileDiff);
    }
  }

  return instances;
}

function cleanupDiffs(instances) {
  for (const inst of instances) {
    try { inst.cleanUp(); } catch (e) {}
  }
}

// Alpine.js stores and components
document.addEventListener('alpine:init', () => {
  // Global store for app-wide state
  Alpine.store('app', {
    darkMode: localStorage.getItem('theme') === 'dark' ||
      (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches),

    init() {
      if (this.darkMode) {
        document.documentElement.classList.add('dark');
      }
    },

    toggleDarkMode() {
      this.darkMode = !this.darkMode;
      document.documentElement.classList.toggle('dark', this.darkMode);
      localStorage.setItem('theme', this.darkMode ? 'dark' : 'light');
    }
  });

  // Context pane store - global so PR cards can access it
  Alpine.store('contextPane', {
    open: false,
    prNumber: null,

    // Header/meta data (populated immediately from PR card)
    prTitle: '',
    prStatusClass: '',
    prUrl: '',
    prState: '',
    prAge: '',
    prAuthorName: '',
    prAuthorUrl: '',
    prScope: '',
    prScopeUrl: '',

    // Body content (loaded async)
    bodyHtml: '',
    loadingBody: false,

    // Diffs
    diffInstances: [],
    loadingDiffs: false,

    // Internal state
    _abortController: null,
    _owner: '',
    _repo: '',
    _stashedDiffs: [],  // Diffs stashed when closing, reused when opening same PR

    async openPane(prNumber, dataset = {}) {
      // Toggle if same PR - no need to fetch, just toggle visibility
      if (prNumber === this.prNumber) {
        if (this.open) {
          // Closing - stash diffs for quick reopen
          this._stashDiffs();
          this.open = false;
        } else {
          // Opening - use stashed diffs if available, otherwise just open
          this._openWithDeferredDiffs();
        }
        return;
      }

      // Different PR - clean up previous pane
      this._cleanup();

      // Set PR number
      this.prNumber = prNumber;

      // Populate header/meta immediately from dataset
      this.prTitle = dataset.prTitle || 'Pull Request';
      this.prStatusClass = dataset.prStatusClass || '';
      this.prUrl = dataset.prUrl || '#';
      this.prState = dataset.prState || '';
      this.prAge = dataset.prAge || '';
      this.prAuthorName = dataset.prAuthorName || '';
      this.prAuthorUrl = dataset.prAuthorUrl || '#';
      this.prScope = dataset.prScope || '';
      this.prScopeUrl = dataset.prScopeUrl || '#';

      // Find owner/repo
      const card = document.querySelector(`.pull-request[data-pr-number="${prNumber}"]`);
      const owner = card?.closest('[data-owner]')?.dataset.owner;
      const repo = card?.closest('[data-repo]')?.dataset.repo;

      if (!owner || !repo) {
        const pathParts = window.location.pathname.split('/').filter(Boolean);
        if (pathParts.length >= 2) {
          this._owner = pathParts[0];
          this._repo = pathParts[1];
        }
      } else {
        this._owner = owner;
        this._repo = repo;
      }

      // Show skeleton for body and diffs
      this.bodyHtml = this._bodySkeletonHtml();
      this.loadingBody = true;
      this.loadingDiffs = true;

      // Open pane immediately with header/meta
      this.open = true;

      // Mark card as selected
      const selected = document.querySelector('.pull-request.selected');
      if (selected) selected.classList.remove('selected');
      card?.classList.add('selected');

      // Fire parallel requests, wait for both, then swap together
      this._fetchBodyAndDiffs(prNumber);
    },

    closePane() {
      this._stashDiffs();
      this._cleanup();

      this.open = false;
      // Keep all data so same-PR reopen shows instantly without refetching
    },

    openPrRef(event) {
      const link = event.target.closest('.pr-ref');
      if (!link) return;
      event.preventDefault();

      const prNumber = parseInt(link.dataset.prNumber, 10);
      if (isNaN(prNumber)) return;

      // Card on page — safe to swap immediately via openPane
      const card = document.querySelector(`.pull-request[data-pr-number="${prNumber}"]`);
      if (card) {
        this.openPane(prNumber, card.dataset);
        return;
      }

      // Card not on page — fetch first, only swap on success
      const pathParts = window.location.pathname.split('/').filter(Boolean);
      const owner = pathParts[0] || this._owner;
      const repo = pathParts[1] || this._repo;
      if (!owner || !repo) return;

      link.classList.add('loading');

      this._fetchBodyRaw(prNumber).then(html => {
        link.classList.remove('loading');

        this._cleanup();
        this.prNumber = prNumber;
        this.prTitle = 'Pull Request #' + prNumber;
        this.prStatusClass = '';
        this.prUrl = `https://github.com/${owner}/${repo}/pull/${prNumber}`;
        this.prState = '';
        this.prAge = '';
        this.prAuthorName = '';
        this.prAuthorUrl = '#';
        this.prScope = '';
        this.prScopeUrl = '#';
        this._owner = owner;
        this._repo = repo;

        this.bodyHtml = html;
        this.open = true;
      }).catch(() => {
        link.classList.remove('loading');
      });
    },

    _stashDiffs() {
      const container = document.getElementById('context-pane-content');
      const body = container?.querySelector('.context-pane-body');
      if (!body) return;

      // Detach diffs and stash them
      body.querySelectorAll('diffs-container').forEach(el => {
        el.remove();
        this._stashedDiffs.push(el);
      });
    },

    _unstashDiffs() {
      if (this._stashedDiffs.length === 0) return false;

      const container = document.getElementById('context-pane-content');
      const body = container?.querySelector('.context-pane-body');
      if (!body) return false;

      // Re-insert stashed diffs
      this._stashedDiffs.forEach(el => body.appendChild(el));
      this._stashedDiffs = [];
      return true;
    },

    _cleanup() {
      // Abort previous requests
      if (this._abortController) {
        this._abortController.abort();
        this._abortController = null;
      }

      // Cleanup diffs
      if (this.diffInstances.length) {
        cleanupDiffs(this.diffInstances);
        this.diffInstances = [];
      }

      // Clear stash (different PR, don't reuse old diffs)
      this._stashedDiffs = [];

      // Also remove any diff elements from DOM
      const container = document.getElementById('context-pane-content');
      if (container) {
        container.querySelectorAll('diffs-container').forEach(el => el.remove());
      }

      // Remove selection from card
      const selected = document.querySelector('.pull-request.selected');
      if (selected) selected.classList.remove('selected');
    },

    async _fetchBodyAndDiffs(prNumber) {
      this._abortController = new AbortController();
      const signal = this._abortController.signal;

      const container = document.getElementById('context-pane-content');
      const target = container?.querySelector('.context-pane-body') || container;

      // Show diff skeleton
      if (target) {
        target.querySelectorAll('diffs-container').forEach(el => el.remove());
        this._showDiffSkeleton(target);
      }

      // Use a temporary container for diffs so they don't render into DOM yet
      const tempDiffContainer = document.createElement('div');

      // Fire both requests in parallel
      const bodyPromise = this._fetchBodyRaw(prNumber, signal);
      const diffsPromise = this._fetchDiffsRaw(prNumber, signal, tempDiffContainer);

      try {
        const [bodyHtml, diffInstances] = await Promise.all([bodyPromise, diffsPromise]);

        if (signal.aborted) return;

        // Remove skeleton and any old diffs
        if (target) {
          target.querySelectorAll('.diff-skeleton, diffs-container').forEach(el => el.remove());
          // Move diffs from temp container to real target
          Array.from(tempDiffContainer.children).forEach(el => target.appendChild(el));
        }

        // Swap in body at the same time
        this.bodyHtml = bodyHtml;
        this.diffInstances = diffInstances;
        this.loadingBody = false;
        this.loadingDiffs = false;

        // Load images after DOM updates - use setTimeout to ensure DOM is rendered
        setTimeout(() => {
          this._loadImages();
        }, 0);
      } catch (err) {
        if (err.name === 'AbortError') return;
        console.error('[fetch]', err);
        this.bodyHtml = '<div style="color: var(--clr-text-muted); font-size: 0.875rem; text-align: center;">Failed to load content</div>';
        this.loadingBody = false;
        this.loadingDiffs = false;
        if (target) {
          target.querySelectorAll('.diff-skeleton').forEach(el => el.remove());
        }
      }
    },

    async _fetchBodyRaw(prNumber, signal) {
      const response = await fetch(`/${this._owner}/${this._repo}?part=pr-detail&number=${prNumber}`, { signal });
      if (!response.ok) throw new Error('Failed to fetch body');

      const html = await response.text();

      // Parse the HTML to extract just the body content
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');
      const bodyElement = doc.querySelector('.context-pane-body');

      if (bodyElement) {
        const meta = bodyElement.querySelector('.context-pane-meta');
        let bodyContent = '';
        let node = bodyElement.firstChild;
        while (node) {
          if (node !== meta && !(node.classList && node.classList.contains('context-pane-meta'))) {
            bodyContent += node.outerHTML || node.textContent;
          }
          node = node.nextSibling;
        }
        return bodyContent;
      }
      return html;
    },

    async _fetchDiffsRaw(prNumber, signal, target) {
      if (!this._owner || !this._repo) return [];
      // Render diffs into temporary container (not visible yet)
      const instances = await loadPRDiffs(target, this._owner, this._repo, prNumber, signal);
      return instances;
    },

    _openWithDeferredDiffs() {
      const container = document.getElementById('context-pane-content');
      const body = container?.querySelector('.context-pane-body');
      const pane = document.getElementById('context-pane');

      // Always show skeleton during transition
      this._showDiffSkeleton(body);

      // Open the pane (animation happens over 50ms)
      this.open = true;

      // After animation, replace skeleton with content
      const hasStashed = this._stashedDiffs.length > 0;
      const replaceSkeleton = () => {
        if (body && this.open) {
          const skeleton = body.querySelector('.diff-skeleton');
          if (skeleton) skeleton.remove();

          if (hasStashed) {
            this._unstashDiffs();
          }
          // If no stashed diffs, skeleton stays until _fetchDiffs completes
        }
        // Re-process images for skeleton loading effect
        setTimeout(() => {
          this._loadImages();
        }, 0);
      };

      this._onTransitionComplete(pane, replaceSkeleton);
    },

    _onTransitionComplete(pane, callback) {
      let handled = false;

      const onTransitionEnd = (e) => {
        if (handled || (pane && e.target !== pane)) return;
        handled = true;
        if (pane) pane.removeEventListener('transitionend', onTransitionEnd);
        callback();
      };

      if (pane) {
        pane.addEventListener('transitionend', onTransitionEnd);
      }

      // Fallback: if transitionend doesn't fire within 100ms, execute anyway
      setTimeout(() => {
        if (!handled) {
          handled = true;
          if (pane) pane.removeEventListener('transitionend', onTransitionEnd);
          callback();
        }
      }, 100);
    },

    _showDiffSkeleton(body) {
      if (!body) return;
      const skeleton = document.createElement('div');
      skeleton.className = 'diff-skeleton skeleton';
      skeleton.style.cssText = 'display: flex; flex-direction: column; gap: 0.75rem;';

      // Create full-width diff-like blocks with varying heights
      const heights = [100, 140, 80, 120, 160];
      for (const h of heights) {
        const block = document.createElement('div');
        block.className = 'skeleton-line';
        block.style.cssText = `
          width: 100%;
          height: ${h}px;
          background-color: var(--clr-decorative);
          border-radius: 4px;
        `;
        skeleton.appendChild(block);
      }
      body.appendChild(skeleton);
    },

    _loadImages() {
      const container = document.getElementById('context-pane-content');
      if (!container) return;
      container.querySelectorAll('img[data-src]').forEach(img => {
        const src = img.dataset.src;
        const preloadImg = new Image();
        preloadImg.onload = () => {
          img.src = src;
          img.removeAttribute('data-src');
          img.classList.add('loaded');
        };
        preloadImg.src = src;
      });
    },

    _bodySkeletonHtml() {
      return `
        <div class="body-skeleton skeleton">
          <div class="skeleton-paragraph">
            <div class="skeleton-line"></div>
            <div class="skeleton-line"></div>
            <div class="skeleton-line"></div>
            <div class="skeleton-line"></div>
            <div class="skeleton-line short"></div>
          </div>
          <div class="skeleton-paragraph">
            <div class="skeleton-line"></div>
            <div class="skeleton-line"></div>
            <div class="skeleton-line"></div>
            <div class="skeleton-line medium"></div>
          </div>
          <div class="skeleton-paragraph">
            <div class="skeleton-line"></div>
            <div class="skeleton-line"></div>
            <div class="skeleton-line short"></div>
          </div>
        </div>
      `;
    }
  });

  // Filter panel store
  Alpine.store('filterPanel', {
    open: false,

    openPanel() {
      this._moveFiltersToPanel();
      this.open = true;
      document.body.classList.add('filter-pane-open');
    },

    close() {
      this.open = false;
      document.body.classList.remove('filter-pane-open');
    },

    _moveFiltersToPanel() {
      const panel = document.querySelector('.filter-panel-body');
      if (!panel) return;

      const scopesSection = document.getElementById('scopes-section');
      const contribsSection = document.getElementById('contributors-section');

      if (scopesSection && scopesSection.parentElement !== panel) {
        panel.appendChild(scopesSection);
      }
      if (contribsSection && contribsSection.parentElement !== panel) {
        panel.appendChild(contribsSection);
      }
    }
  });

  // Nav component with title inputs
  Alpine.data('nav', () => ({
    origOwner: '',
    origRepo: '',
    origScope: '',
    _fontsReady: false,

    init() {
      // Wait for fonts before first resize
      document.fonts.ready.then(() => {
        this._fontsReady = true;
        this._resizeAllInputs();
      });

      this.$nextTick(() => {
        this._captureOriginals();
        this._setupInputs();
      });

      // Re-setup inputs after HTMX swap
      document.addEventListener('htmx:afterSwap', () => {
        this.$nextTick(() => {
          this._captureOriginals();
          this._setupInputs();
          this._syncScopeFromUrl();
          // Resize after fonts are ready
          if (this._fontsReady) {
            this._resizeAllInputs();
          } else {
            document.fonts.ready.then(() => this._resizeAllInputs());
          }
        });
      });
    },

    _captureOriginals() {
      this.origOwner = document.querySelector('.title-input[data-part="owner"]')?.value.trim() || '';
      this.origRepo = document.querySelector('.title-input[data-part="repo"]')?.value.trim() || '';
      this.origScope = document.querySelector('.title-input[data-part="scope"]')?.value.trim() || '';
    },

    _setupInputs() {
      const inputs = document.querySelectorAll('.title-input');
      inputs.forEach(input => {
        this._resize(input);
        // Remove old listeners by cloning (prevents duplicates)
        const newInput = input.cloneNode(true);
        input.parentNode.replaceChild(newInput, input);

        newInput.addEventListener('input', () => this._resize(newInput));
        newInput.addEventListener('change', () => this._resize(newInput));
        newInput.addEventListener('focus', () => newInput.select());
        newInput.addEventListener('keydown', (ev) => {
          if (ev.key === 'Enter' && newInput.hasAttribute('data-part')) {
            ev.preventDefault();
            this._navigate();
          }
        });
        newInput.addEventListener('blur', () => {
          setTimeout(() => {
            if (document.activeElement?.classList.contains('title-input')) return;
            this._navigate();
          }, 0);
        });
      });
    },

    _resize(input) {
      input.style.width = '1px';
      if (input.value) {
        input.style.width = input.scrollWidth + 'px';
      } else if (input.placeholder) {
        const prev = input.value;
        input.value = input.placeholder;
        input.style.width = input.scrollWidth + 'px';
        input.value = prev;
      }
    },

    _resizeAllInputs() {
      document.querySelectorAll('.title-input').forEach(input => this._resize(input));
    },

    _navigate() {
      const owner = (document.querySelector('.title-input[data-part="owner"]')?.value || '').trim();
      const repo = (document.querySelector('.title-input[data-part="repo"]')?.value || '').trim();
      const scope = (document.querySelector('.title-input[data-part="scope"]')?.value || '').trim();

      if (owner === this.origOwner && repo === this.origRepo && scope === this.origScope) return;

      if (repo) {
        let url = `/${owner}/${repo}`;
        if (owner === this.origOwner && repo === this.origRepo && scope !== this.origScope) {
          if (scope) {
            url += `/${scope}`;
          }
        }
        htmx.ajax('GET', url, {
          target: 'main',
          swap: 'innerHTML'
        });
      }
    },

    _syncScopeFromUrl() {
      const parts = window.location.pathname.split('/').filter(Boolean);
      if (parts.length > 2) {
        const curScope = parts.slice(2).join('/');
        if (curScope !== this.origScope) {
          const inp = document.querySelector('.title-input[data-part="scope"]');
          if (inp) {
            inp.value = curScope;
            this.origScope = curScope;
            inp.dispatchEvent(new Event('input'));
          }
        }
      } else {
        const inp = document.querySelector('.title-input[data-part="scope"]');
        if (inp && inp.value) {
          inp.value = '';
          this.origScope = '';
          inp.dispatchEvent(new Event('input'));
        }
      }
    }
  }));
});
