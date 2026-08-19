// Durable-save layer injected by the Windows wrapper.
// It serializes snapshots, keeps a recovery copy, and prevents a stale
// auto-save from replacing a newer manual save.
(function(){
  const CACHE_KEY='amit_patel_dental_save_cache_v1';
  let version=0;
  let queue=Promise.resolve();

  function cloneDB(){
    try{return JSON.parse(JSON.stringify(db));}catch(e){return null;}
  }
  function cacheSnapshot(payload){try{localStorage.setItem(CACHE_KEY,JSON.stringify(payload));}catch(e){}}
  function readCache(){try{return JSON.parse(localStorage.getItem(CACHE_KEY)||'null');}catch(e){return null;}}

  window.persistDB=function(){
    const payload=cloneDB();
    if(!payload)return Promise.resolve();
    const next=Math.max(++version,Date.now());
    payload._meta={version:next,savedAt:new Date().toISOString()};
    cacheSnapshot(payload);
    if(!serverOnline)return Promise.resolve();

    queue=queue.then(async()=>{
      try{
        const r=await fetch('/api/db',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload),keepalive:true});
        if(!r.ok)throw new Error('Save failed');
      }catch(e){
        serverOnline=false;
        const st=document.getElementById('serverStatus');
        if(st)st.textContent='● Database connection lost — local recovery copy retained';
      }
    });
    return queue;
  };

  const originalLoadServerDB=loadServerDB;
  window.loadServerDB=async function(){
    const cached=readCache();
    try{await originalLoadServerDB();}catch(e){}
    const cacheVersion=Number(cached?._meta?.version||0);
    const serverVersion=Number(db?._meta?.version||0);
    if(cached && cacheVersion>serverVersion && Array.isArray(cached.patients) && Array.isArray(cached.visits) && Array.isArray(cached.payments)){
      normalizeDB(cached);
      version=cacheVersion;
      serverOnline=true;
      await persistDB();
    }else if(serverVersion>0){
      version=serverVersion;
    }
  };

  window.saveDB=function(){
    try{ensureReceiptNumbers();}catch(e){}
    persistDB();
    try{renderAll();}catch(e){}
    try{updateSuggestions();}catch(e){}
  };

  window.addEventListener('beforeunload',function(){try{persistDB();}catch(e){}});
})();
