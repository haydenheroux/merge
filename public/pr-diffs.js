let diffsLib = null;

async function ensureLoaded() {
  if (diffsLib) return diffsLib;
  diffsLib = await import('https://esm.sh/@pierre/diffs@1.2.12');
  return diffsLib;
}

function getTheme() {
  return document.documentElement.classList.contains('dark') ? 'pierre-dark' : 'pierre-light';
}

export async function loadPRDiffs(container, owner, repo, prNumber, signal) {
  const lib = await ensureLoaded();

  const resp = await fetch(`/${owner}/${repo}/pull/${prNumber}/diff`, { signal });
  if (!resp.ok) throw new Error('Failed to fetch diff: ' + resp.status);

  const diffText = await resp.text();
  const patches = lib.parsePatchFiles(diffText);

  const instances = [];

  for (const patch of patches) {
    for (const fileDiffMeta of patch.files) {
      const fileDiff = new lib.FileDiff({
        theme: getTheme(),
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

export function cleanupDiffs(instances) {
  for (const inst of instances) {
    try { inst.cleanUp(); } catch (e) {}
  }
}
