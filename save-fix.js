(() => {
  const RECOVERY_KEY = 'amit_patel_dental_recovery_v2';
  let writeChain = Promise.resolve();
  let version = 0;

  const clone = value => JSON.parse(JSON.stringify(value));
  const setStatus = text => { const el = document.getElementById('serverStatus'); if (el) el.textContent = text; };

  function cacheSnapshot(snapshot) {
    try {
      localStorage.setItem(RECOVERY_KEY, JSON.stringify({
        version,
        savedAt: new Date().toISOString(),
        data: snapshot
      }));
    } catch (_) {}
  }

  function readSnapshot() {
    try { return JSON.parse(localStorage.getItem(RECOVERY_KEY) || 'null'); } catch (_) { return null; }
  }

  window.persistDB = function () {
    let snapshot;
    try { snapshot = clone(window.db); } catch (_) { return writeChain; }
    version = Math.max(version + 1, Date.now());
    snapshot._meta = { version, savedAt: new Date().toISOString() };
    cacheSnapshot(snapshot);

    writeChain = writeChain.then(async () => {
      try {
        const r = await fetch('/api/db', {
          method: 'PUT',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(snapshot),
          cache: 'no-store',
          keepalive: true
        });
        if (!r.ok) throw new Error('save failed');
        setStatus('● Local database connected');
      } catch (_) {
        cacheSnapshot(snapshot);
        setStatus('● Saved locally; database retrying');
      }
    });
    return writeChain;
  };

  // All existing saveDB() calls in the master HTML continue to work.
  window.saveDB = function () {
    try { if (typeof ensureReceiptNumbers === 'function') ensureReceiptNumbers(); } catch (_) {}
    window.persistDB();
    try { if (typeof renderAll === 'function') renderAll(); } catch (_) {}
    try { if (typeof updateSuggestions === 'function') updateSuggestions(); } catch (_) {}
  };

  // Restore a newer recovery snapshot if the database contains an older snapshot.
  const originalLoad = window.loadServerDB;
  window.loadServerDB = async function () {
    const cached = readSnapshot();
    try { await originalLoad(); } catch (_) {}
    const cachedVersion = Number(cached?.version || 0);
    const serverVersion = Number(window.db?._meta?.version || 0);
    if (cached?.data && cachedVersion > serverVersion && Array.isArray(cached.data.patients) && Array.isArray(cached.data.visits) && Array.isArray(cached.data.payments)) {
      normalizeDB(cached.data);
      version = cachedVersion;
      setStatus('● Newer local recovery copy loaded');
      await window.persistDB();
    } else if (serverVersion) {
      version = serverVersion;
    }
  };

  // Replace UTC-based date calculation with the Windows computer's local calendar date.
  window.todayISO = function () {
    const d = new Date();
    const p = n => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${p(d.getMonth()+1)}-${p(d.getDate())}`;
  };

  // Recovery checkpoint every 5 seconds. Database writes are serialized above.
  setInterval(() => { try { window.persistDB(); } catch (_) {} }, 5000);

  window.addEventListener('beforeunload', () => {
    try {
      const snapshot = clone(window.db);
      snapshot._meta = { version: Math.max(++version, Date.now()), savedAt: new Date().toISOString() };
      cacheSnapshot(snapshot);
      navigator.sendBeacon('/api/db-beacon', new Blob([JSON.stringify(snapshot)], {type:'application/json'}));
    } catch (_) {}
  });
})();
