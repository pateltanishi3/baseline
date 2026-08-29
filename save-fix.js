// Durable local-save layer. Persistence is deliberately independent from UI rendering.
(function(){
  'use strict';
  const CACHE_KEY='amit_patel_dental_save_cache_v1';
  let version=0, queue=Promise.resolve(), saving=false;
  function cloneDB(){try{return JSON.parse(JSON.stringify(db));}catch(e){return null;}}
  function cacheSnapshot(p){try{localStorage.setItem(CACHE_KEY,JSON.stringify(p));}catch(e){}}
  function readCache(){try{return JSON.parse(localStorage.getItem(CACHE_KEY)||'null');}catch(e){return null;}}
  window.persistDB=function(){
    const payload=cloneDB(); if(!payload)return Promise.resolve();
    payload._meta={version:Math.max(++version,Date.now()),savedAt:new Date().toISOString()};
    version=payload._meta.version; cacheSnapshot(payload);
    if(!serverOnline||saving)return Promise.resolve();
    queue=queue.then(async()=>{saving=true;try{const r=await fetch('/api/db',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});if(!r.ok)throw new Error('Save failed');}catch(e){serverOnline=false;const st=document.getElementById('serverStatus');if(st)st.textContent='● Database connection lost — local recovery copy retained';}finally{saving=false;}}); return queue;
  };
  const originalLoadServerDB=window.loadServerDB;
  if(typeof originalLoadServerDB==='function') window.loadServerDB=async function(){
    const cached=readCache(); try{await originalLoadServerDB();}catch(e){}
    const cv=Number(cached?._meta?.version||0),sv=Number(db?._meta?.version||0);
    if(cached&&cv>sv&&Array.isArray(cached.patients)&&Array.isArray(cached.visits)&&Array.isArray(cached.payments)){
      normalizeDB(cached);version=cv;serverOnline=true;
      const payload=cloneDB();if(payload)await fetch('/api/db',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)}).catch(()=>{});
    }else if(sv>0)version=sv;
  };
  window.saveDB=function(){try{ensureReceiptNumbers();}catch(e){} return persistDB();};
  window.todayISO=function(){const d=new Date(),y=d.getFullYear(),m=String(d.getMonth()+1).padStart(2,'0'),day=String(d.getDate()).padStart(2,'0');return `${y}-${m}-${day}`;};
  window.addEventListener('beforeunload',()=>{try{persistDB();}catch(e){}});
})();
