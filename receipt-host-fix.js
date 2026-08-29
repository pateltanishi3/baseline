(function(){
  'use strict';
  // The receipt-settings module expects a #settings host. Build #7's HTML does not
  // expose that id, so locate the existing Prescription Size settings card and use
  // its containing settings panel as the host before receipt-settings boots.
  function findHost(){
    if(document.getElementById('settings')) return document.getElementById('settings');
    const cards=Array.from(document.querySelectorAll('.card'));
    const rx=cards.find(c=>/Prescription Size/i.test(c.textContent||''));
    if(!rx) return null;
    const host=rx.parentElement || rx;
    if(!host.id) host.id='settings';
    return host;
  }
  function boot(){ try{ findHost(); }catch(e){ console.error('Receipt settings host fix:',e); } }
  boot();
  setTimeout(boot,50); setTimeout(boot,200); setTimeout(boot,500);
})();
