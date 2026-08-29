(function(){
  'use strict';
  const DEFAULT={width:145,height:190,top:0,right:0,bottom:0,left:0,fontFamily:'Arial',fontSize:10};
  function settings(){
    db.settings=db.settings||{};
    db.settings.receiptPdf=Object.assign({},DEFAULT,db.settings.receiptPdf||{});
    return db.settings.receiptPdf;
  }
  function save(){ if(typeof saveDB==='function') saveDB(); }
  function money(v){return Number(v||0).toLocaleString('en-IN');}
  function esc(v){return String(v==null?'':v).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');}
  function mm(v){return Number(v||0)*2.834645669;}
  function pdfEscape(v){return String(v==null?'':v).replace(/\\/g,'\\\\').replace(/\(/g,'\\(').replace(/\)/g,'\\)').replace(/[^\x20-\x7E]/g,'?');}
  function wrap(text,max){
    const words=String(text||'—').split(/\s+/), out=[]; let line='';
    words.forEach(w=>{if(!line){line=w;return} if((line+' '+w).length<=max) line+=' '+w; else {out.push(line);line=w;}});
    if(line)out.push(line); return out.length?out:['—'];
  }
  function fontMap(name){
    if(name==='Times New Roman')return 'Times-Roman';
    if(name==='Courier New')return 'Courier';
    return 'Helvetica';
  }
  function makePdf(patient,payment){
    const s=settings();
    const W=Math.max(40,Number(s.width)||145), H=Math.max(40,Number(s.height)||190);
    const T=Math.max(0,Number(s.top)||0), R=Math.max(0,Number(s.right)||0), B=Math.max(0,Number(s.bottom)||0), L=Math.max(0,Number(s.left)||0);
    const fs=Math.min(24,Math.max(7,Number(s.fontSize)||10));
    const ptW=mm(W),ptH=mm(H), left=mm(L), right=mm(R), top=mm(T), bottom=mm(B);
    const printableW=Math.max(20,ptW-left-right);
    const visit=(db.visits||[]).filter(v=>v.patientId===patient.id&&v.date===payment.date).sort((a,b)=>String(b.id||'').localeCompare(String(a.id||'')))[0]||{};
    const receiptNo=payment.receiptNo||payment.id;
    const rows=[
      ['Date',payment.date||'—'],['Receipt No.',receiptNo],['Patient ID',patient.id||'—'],['Patient Name',patient.name||'—'],['Mobile',patient.phone||'—'],
      ['Diagnosis',visit.diagnosis||'—'],['Treatment',payment.treatment||visit.treatment||'—'],['Payment Mode',payment.mode||'—'],
      ['Total Treatment Cost','₹ '+money(payment.cost)],['Payment Received','₹ '+money(payment.paid)],['Balance','₹ '+money((Number(payment.cost)||0)-(Number(payment.paid)||0))]
    ];
    const lines=[]; const bodyFont=fontMap(s.fontFamily);
    let y=ptH-top-24;
    const lineH=fs*1.55;
    function text(txt,x,yy,size,bold){lines.push(`BT /${bold?'F2':'F1'} ${size} Tf ${x.toFixed(2)} ${yy.toFixed(2)} Td (${pdfEscape(txt)}) Tj ET`);}
    text('PAYMENT RECEIPT',left,y,fs+4,true); y-=lineH*1.8;
    const labelW=85;
    rows.forEach(([label,value])=>{
      const chunks=wrap(value,Math.max(18,Math.floor((printableW-labelW-10)/(fs*4.2))));
      text(label,left,y,fs,true); text(chunks[0],left+labelW,y,fs,false); y-=lineH;
      for(let i=1;i<chunks.length;i++){text(chunks[i],left+labelW,y,fs,false);y-=lineH;}
      y-=fs*0.35;
    });
    const content=new TextEncoder().encode(lines.join('\n')+'\n');
    const objs=[];
    objs[1]=new TextEncoder().encode('<< /Type /Catalog /Pages 2 0 R >>');
    objs[2]=new TextEncoder().encode('<< /Type /Pages /Kids [3 0 R] /Count 1 >>');
    objs[3]=new TextEncoder().encode(`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${ptW.toFixed(2)} ${ptH.toFixed(2)}] /Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>`);
    objs[4]=new TextEncoder().encode(`<< /Type /Font /Subtype /Type1 /BaseFont /${bodyFont} >>`);
    objs[5]=new TextEncoder().encode(`<< /Type /Font /Subtype /Type1 /BaseFont /${bodyFont==='Times-Roman'?'Times-Bold':bodyFont==='Courier'?'Courier-Bold':'Helvetica-Bold'} >>`);
    objs[6]=new TextEncoder().encode(`<< /Length ${content.length} >>\nstream\n`); objs[7]=content; objs[8]=new TextEncoder().encode('endstream');
    const chunks=[new TextEncoder().encode('%PDF-1.4\n%âãÏÓ\n')]; const offsets=[0]; let pos=chunks[0].length;
    for(let i=1;i<=8;i++){offsets[i]=pos; const head=new TextEncoder().encode(`${i} 0 obj\n`); chunks.push(head);pos+=head.length;chunks.push(objs[i]);pos+=objs[i].length;const tail=new TextEncoder().encode('\nendobj\n');chunks.push(tail);pos+=tail.length;}
    const xref=pos; let xr='xref\n0 9\n0000000000 65535 f \n'; for(let i=1;i<=8;i++)xr+=String(offsets[i]).padStart(10,'0')+' 00000 n \n';
    chunks.push(new TextEncoder().encode(xr)); chunks.push(new TextEncoder().encode(`trailer\n<< /Size 9 /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF`));
    return new Blob(chunks,{type:'application/pdf'});
  }
  window.makeReceipt=window.makeReceipt||function(paymentId){
    const payment=(db.payments||[]).find(x=>x.id===paymentId); if(!payment){alert('Payment receipt not found.');return;}
    const patient=(db.patients||[]).find(x=>x.id===payment.patientId); if(!patient){alert('Patient not found.');return;}
    const blob=makePdf(patient,payment), a=document.createElement('a'); a.href=URL.createObjectURL(blob); a.download='Receipt_'+(payment.receiptNo||payment.id)+'.pdf'; document.body.appendChild(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(a.href),2000);
  };
  function ensureCard(){
    const host=document.getElementById('settings'); if(!host||document.getElementById('receiptPdfSettings'))return;
    const s=settings(), card=document.createElement('div'); card.className='card'; card.id='receiptPdfSettings'; card.style.marginTop='16px';
    card.innerHTML=`<h2 style="margin-top:0">🧾 Payment Receipt PDF Settings</h2><p class="small">Same physical-size and print-spacing controls as the Prescription settings. These values are applied to the Payment Receipt PDF.</p><div class="formgrid"><div class="field"><label>Width (cm)</label><input id="rpWidth" type="number" min="4" max="100" step="0.1"></div><div class="field"><label>Height (cm)</label><input id="rpHeight" type="number" min="4" max="100" step="0.1"></div><div class="field"><label>Font</label><select id="rpFont"><option>Arial</option><option>Segoe UI</option><option>Calibri</option><option>Times New Roman</option><option>Courier New</option></select></div><div class="field"><label>Font Size (pt)</label><input id="rpFontSize" type="number" min="7" max="24" step="0.5"></div></div><h3 style="margin:18px 0 8px">🖨️ Payment Receipt Print Spacing</h3><div class="formgrid"><div class="field"><label>Top Height (mm)</label><input id="rpTop" type="number" min="0" step="0.5"></div><div class="field"><label>Right Side Width (mm)</label><input id="rpRight" type="number" min="0" step="0.5"></div><div class="field"><label>Bottom Height (mm)</label><input id="rpBottom" type="number" min="0" step="0.5"></div><div class="field"><label>Left Side Width (mm)</label><input id="rpLeft" type="number" min="0" step="0.5"></div></div><div class="actions"><button type="button" class="primary" id="rpSave">Save Receipt PDF Settings</button><button type="button" class="secondary" id="rpReset">Reset Receipt Settings</button></div>`;
    host.appendChild(card);
    const load=()=>{const x=settings();rpWidth.value=(Number(x.width)/10).toFixed(1);rpHeight.value=(Number(x.height)/10).toFixed(1);rpFont.value=x.fontFamily||'Arial';rpFontSize.value=x.fontSize||10;rpTop.value=x.top||0;rpRight.value=x.right||0;rpBottom.value=x.bottom||0;rpLeft.value=x.left||0;};
    rpSave.onclick=()=>{const w=Number(rpWidth.value),h=Number(rpHeight.value),fs=Number(rpFontSize.value);const vals=[rpTop,rpRight,rpBottom,rpLeft].map(e=>Number(e.value));if(!w||!h||w>100||h>100){alert('Enter a valid receipt width and height.');return}if(!fs||fs<7||fs>24){alert('Font size must be between 7 and 24 pt.');return}if(vals.some(v=>!Number.isFinite(v)||v<0)||vals[0]+vals[2]>=h*10||vals[1]+vals[3]>=w*10){alert('Receipt spacing is too large for the selected paper size.');return}db.settings.receiptPdf={width:w*10,height:h*10,top:vals[0],right:vals[1],bottom:vals[2],left:vals[3],fontFamily:rpFont.value,fontSize:fs};save();load();alert('Payment Receipt PDF settings saved.');};
    rpReset.onclick=()=>{db.settings.receiptPdf=Object.assign({},DEFAULT);save();load();}; load();
  }
  const boot=()=>{try{ensureCard();}catch(e){console.error(e);}};
  boot(); setTimeout(boot,300); setTimeout(boot,1000); setInterval(boot,1500);
})();
