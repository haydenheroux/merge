// Diff library loading
let diffsLib = null;

async function ensureDiffsLib() {
  if (diffsLib) return diffsLib;
  diffsLib = await import('https://esm.sh/@pierre/diffs@1.2.12');
  return diffsLib;
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

    async openPane(prNumber, dataset) {
      // Toggle if same PR
      if (prNumber === this.prNumber && this.open) {
        this.closePane();
        return;
      }

      // Clean up previous pane
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

      // Show skeleton for body
      this.bodyHtml = this._bodySkeletonHtml();
      this.loadingBody = true;
      this.loadingDiffs = true;

      // Open pane immediately with header/meta
      this.open = true;

      // Mark card as selected
      const selected = document.querySelector('.pull-request.selected');
      if (selected) selected.classList.remove('selected');
      card?.classList.add('selected');

      // Fire parallel requests for body and diffs
      this._fetchBody(prNumber);
      this._fetchDiffs(prNumber);
    },

    closePane() {
      this._cleanup();
      
      this.open = false;
      this.prNumber = null;
      this.prTitle = '';
      this.prStatusClass = '';
      this.prUrl = '';
      this.prState = '';
      this.prAge = '';
      this.prAuthorName = '';
      this.prAuthorUrl = '';
      this.prScope = '';
      this.prScopeUrl = '';
      this.bodyHtml = '';
      this.loadingBody = false;
      this.loadingDiffs = false;
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

      // Also remove any diff elements from DOM
      const container = document.getElementById('context-pane-content');
      if (container) {
        container.querySelectorAll('diffs-container').forEach(el => el.remove());
      }

      // Remove selection from card
      const selected = document.querySelector('.pull-request.selected');
      if (selected) selected.classList.remove('selected');
    },

    async _fetchBody(prNumber) {
      try {
        this._abortController = new AbortController();
        const response = await fetch(`/${this._owner}/${this._repo}?part=pr-detail&number=${prNumber}`, {
          signal: this._abortController.signal
        });

        if (!response.ok) throw new Error('Failed to fetch');

        const html = await response.text();
        if (this._abortController.signal.aborted) return;

        // Parse the HTML to extract just the body content
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        const bodyElement = doc.querySelector('.context-pane-body');
        
        if (bodyElement) {
          // Get the content after the meta section
          const meta = bodyElement.querySelector('.context-pane-meta');
          let bodyContent = '';
          let node = bodyElement.firstChild;
          while (node) {
            if (node !== meta && !(node.classList && node.classList.contains('context-pane-meta'))) {
              bodyContent += node.outerHTML || node.textContent;
            }
            node = node.nextSibling;
          }
          this.bodyHtml = bodyContent;
        } else {
          this.bodyHtml = html;
        }
        
        this.loadingBody = false;

        // Load images after DOM updates
        Alpine.nextTick(() => {
          this._loadImages();
        });
      } catch (err) {
        if (err.name === 'AbortError') return;
        console.error('[body]', err);
        this.bodyHtml = '<div style="padding: 1rem; color: var(--clr-text-muted); font-size: 0.875rem; text-align: center;">Failed to load content</div>';
        this.loadingBody = false;
      }
    },

    async _fetchDiffs(prNumber) {
      const container = document.getElementById('context-pane-content');
      if (!container || !this._owner || !this._repo) return;

      const target = container.querySelector('.context-pane-body') || container;
      
      // Clear any existing diff elements from previous PR
      target.querySelectorAll('diffs-container').forEach(el => el.remove());
      
      try {
        const instances = await loadPRDiffs(target, this._owner, this._repo, prNumber, this._abortController?.signal);
        if (!this._abortController?.signal.aborted) {
          this.diffInstances = instances;
        }
      } catch (err) {
        if (err.name === 'AbortError') return;
        console.error('[diffs]', err);
      } finally {
        this.loadingDiffs = false;
      }
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
        <div class="skeleton">
          <div class="skeleton-line" style="width: 90%;"></div>
          <div class="skeleton-line" style="width: 75%;"></div>
          <div class="skeleton-line" style="width: 85%;"></div>
          <div class="skeleton-line" style="width: 60%;"></div>
          <div class="skeleton-line" style="width: 80%;"></div>
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
