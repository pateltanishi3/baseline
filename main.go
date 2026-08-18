package main

import (
    _ "embed"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    webview2 "github.com/jchv/go-webview2"
)

//go:embed Dental_Patient_Management_Software.html
var indexHTML []byte

//go:embed app-icon.png
var appIconPNG []byte

//go:embed app-icon.jpg
var appIconJPEG []byte

var emptyDB = map[string]any{
    "patients": []any{}, "visits": []any{}, "payments": []any{}, "appointments": []any{},
    "suggestions": map[string]any{"complaint": []any{}, "diagnosis": []any{}, "treatment": []any{}, "instructions": []any{}},
    "medicines": []any{}, "prescriptionTemplates": []any{},
    "settings": map[string]any{"theme": "ocean", "rxWidth": 145, "rxHeight": 190, "rxTop": 0, "rxRight": 0, "rxBottom": 0, "rxLeft": 0, "doctorName": "Dr. Amit J. Patel", "clinicName": "Dr. Patel's Dental Implant Center & Clinic"},
}

func dataFile() string {
    base, err := os.UserConfigDir()
    if err != nil { base, _ = os.UserHomeDir() }
    dir := filepath.Join(base, "Amit Patel Dental Software")
    _ = os.MkdirAll(dir, 0700)
    return filepath.Join(dir, "patient-data.json")
}

func ensureDB(p string) {
    if _, err := os.Stat(p); err == nil { return }
    b, _ := json.MarshalIndent(emptyDB, "", "  ")
    _ = os.WriteFile(p, b, 0600)
}

func readDB(p string) []byte {
    b, err := os.ReadFile(p)
    if err != nil { b, _ = json.Marshal(emptyDB) }
    return b
}

func writeDB(p string, r io.Reader) error {
    var v any
    if err := json.NewDecoder(r).Decode(&v); err != nil { return err }
    b, _ := json.MarshalIndent(v, "", "  ")
    tmp := p + ".tmp"
    if err := os.WriteFile(tmp, b, 0600); err != nil { return err }
    return os.Rename(tmp, p)
}

const brandingJS = `(() => {
  const APP_LOGO_DATA = __APP_LOGO_DATA__;
  const APP_LOGO_JPEG = __APP_LOGO_JPEG__;
  const DEFAULT_DOCTOR = 'Dr. Amit J. Patel';
  const DEFAULT_CLINIC = "Dr. Patel's Dental Implant Center & Clinic";
  function brandProfile(){
    db.settings=db.settings||{};
    return { doctor:String(db.settings.doctorName||DEFAULT_DOCTOR).trim()||DEFAULT_DOCTOR, clinic:String(db.settings.clinicName||DEFAULT_CLINIC).trim()||DEFAULT_CLINIC };
  }
  function brandDoctor(){ return brandProfile().doctor; }
  function brandClinic(){ return brandProfile().clinic; }
  function applyBrandLogo(src){
    const use=src&&String(src).startsWith('data:')?src:APP_LOGO_DATA;
    document.querySelectorAll('.brand-logo').forEach(img=>img.src=use);
    const p=document.getElementById('settingsLogoPreview'); if(p)p.src=use;
  }
  if(typeof applyClinicLogo==='function'){
    applyClinicLogo=function(src){applyBrandLogo(src);};
    applyBrandLogo(db.settings?.logo||'');
  }
  function refreshBranding(){
    const p=brandProfile();
    document.title=p.clinic+' — Dental Patient Management Software';
    const d=document.getElementById('dashClinicName'); if(d)d.textContent=p.clinic;
    const g=document.getElementById('dashGreeting'); if(g)g.textContent='Good Morning, '+p.doctor+' 👋';
    applyBrandLogo(db.settings?.logo||'');
  }
  function addBrandSettings(){
    if(document.getElementById('clinicProfileSettings')){refreshBranding();return;}
    const section=document.getElementById('settings'); if(!section)return;
    const card=document.createElement('div'); card.className='card'; card.id='clinicProfileSettings';
    card.innerHTML=`<h2 style="margin-top:0">👨‍⚕️ Doctor & Clinic Profile</h2><p class="small">Set the identity for this software installation. These details are used in the dashboard, greetings, prescriptions, receipts and WhatsApp messages.</p><div class="formgrid"><div class="field"><label>Doctor Name</label><input id="brandDoctorName" placeholder="e.g. Dr. Amit Patel"></div><div class="field"><label>Dental Clinic / Practice Name</label><input id="brandClinicName" placeholder="e.g. Amit Dental Clinic"></div></div><div class="settings-actions"><button type="button" class="primary" id="saveBrandProfileBtn">Save Doctor & Clinic Profile</button><button type="button" class="secondary" id="restoreBrandProfileBtn">Restore Default</button></div><div class="settings-note">Once saved, the doctor/clinic identity replaces the original doctor's name in generated greetings and WhatsApp messages. The logo remains optional and can be changed below.</div>`;
    const firstCard=section.querySelector('.card'); firstCard?.before(card);
    const d=document.getElementById('brandDoctorName'),c=document.getElementById('brandClinicName'); d.value=brandDoctor(); c.value=brandClinic();
    document.getElementById('saveBrandProfileBtn').onclick=()=>{const doctor=d.value.trim(),clinic=c.value.trim();if(!doctor||!clinic){alert('Please enter both Doctor Name and Dental Clinic / Practice Name.');return;}db.settings=db.settings||{};db.settings.doctorName=doctor;db.settings.clinicName=clinic;saveDB();refreshBranding();alert('Doctor and clinic profile saved.');};
    document.getElementById('restoreBrandProfileBtn').onclick=()=>{db.settings=db.settings||{};db.settings.doctorName=DEFAULT_DOCTOR;db.settings.clinicName=DEFAULT_CLINIC;saveDB();d.value=DEFAULT_DOCTOR;c.value=DEFAULT_CLINIC;refreshBranding();};
    refreshBranding();
  }
  function dynamicReceiptFooterJpeg(){
    const canvas=document.createElement('canvas');canvas.width=800;canvas.height=120;const ctx=canvas.getContext('2d');ctx.fillStyle='#fff';ctx.fillRect(0,0,800,120);ctx.strokeStyle='#b52a2a';ctx.lineWidth=3;ctx.beginPath();ctx.moveTo(25,12);ctx.lineTo(775,12);ctx.stroke();ctx.fillStyle='#18364f';ctx.textAlign='center';ctx.font='bold 22px Segoe UI, Arial';ctx.fillText(brandClinic(),400,55);ctx.fillStyle='#4b5563';ctx.font='16px Segoe UI, Arial';ctx.fillText(brandDoctor(),400,82);ctx.fillStyle='#2f8b8c';ctx.font='bold 13px Segoe UI, Arial';ctx.fillText('Dental Patient Management Software',400,104);return canvas.toDataURL('image/jpeg',0.9).split(',')[1];
  }
  if(typeof buildReceiptBlob==='function'){
    const source=buildReceiptBlob.toString().replace(/pdfAscii\('Dr\. Amit J\. Patel'\)/g,'pdfAscii(brandDoctor())').replace(/base64Bytes\(RECEIPT_FOOTER_JPEG\)/g,'base64Bytes(dynamicReceiptFooterJpeg())').replace(/base64Bytes\(RECEIPT_LOGO_JPEG\)/g,'base64Bytes(APP_LOGO_JPEG)');
    try{buildReceiptBlob=eval('('+source+')');}catch(e){console.warn('Dynamic receipt branding unavailable',e);}
  }
  sendWhatsApp=function(){const phone=val('phone'),name=val('name');if(!phone){alert('Enter Contact No. first');return;}const meds=getCurrentMedicines();const rx=meds.length?'\n\nPrescription:\n'+meds.map((m,i)=>`${i+1}. ${m.name} | ${m.dose||'—'} | ${m.frequency||'—'} | ${m.duration||'—'}`).join('\n'):'';const msg=`Dear ${name||'Patient'},\n\nThank you for visiting us.\n\nDiagnosis: ${val('diagnosis')||'—'}\nTreatment: ${val('treatment')||'—'}${rx}\n\n${val('follow')?'Follow-up appointment: '+val('follow')+(val('followTime')?' at '+formatTime(val('followTime')):'')+'\n\n':''}Instructions: ${val('instructions')||'—'}\n\nRegards,\n${brandDoctor()}\n${brandClinic()}`;window.open(waUrl(phone,msg),'_blank');};
  sendPrescriptionWhatsApp=function(visitId){const v=db.visits.find(x=>x.id===visitId);if(!v)return;const p=getPatient(v.patientId);if(!p||!p.phone){alert('Patient has no WhatsApp/contact number.');return;}const meds=v.medicines||[];if(!meds.length){alert('No medicines in this prescription.');return;}const msg=`Dear ${p.name},\n\nPrescription for ${v.date}:\n\n${meds.map((m,i)=>`${i+1}. ${m.name}\n   Dose: ${m.dose||'—'}\n   Frequency: ${m.frequency||'—'}\n   Duration / Instructions: ${m.duration||m.instructions||'—'}`).join('\n\n')}\n\nDiagnosis: ${v.diagnosis||'—'}\nTreatment: ${v.treatment||'—'}\n${v.follow?'Next Follow-up: '+v.follow+(v.followTime?' at '+formatTime(v.followTime):'')+'\n':''}Instructions: ${v.instructions||'—'}\n\nRegards,\n${brandDoctor()}\n${brandClinic()}`;window.open(waUrl(p.phone,msg),'_blank');};
  loadBulkTemplate=function(){const t=document.getElementById('bulkTemplate')?.value,m=document.getElementById('bulkMessage');if(!m)return;const msgs={independence:`🇮🇳 Happy Independence Day! 🇮🇳\n\nWishing you and your family a very Happy Independence Day.\n\nRegards,\n${brandDoctor()}\n${brandClinic()}`,diwali:`🪔 Wishing you and your family a very Happy Diwali!\n\nRegards,\n${brandDoctor()}\n${brandClinic()}`,holiday:`Dear Patient,\n\nOur clinic will remain closed on the mentioned holiday. Please contact us for appointment assistance.\n\nRegards,\n${brandDoctor()}\n${brandClinic()}`,dental:`Dear Patient,\n\nRegular dental check-ups and good oral hygiene help maintain healthy teeth and gums.\n\nRegards,\n${brandDoctor()}\n${brandClinic()}`};m.value=msgs[t]||'';};
  sendReceiptWhatsApp=async function(paymentId){const payment=db.payments.find(x=>x.id===paymentId);if(!payment){alert('Payment receipt not found.');return;}const patient=db.patients.find(x=>x.id===payment.patientId);if(!patient){alert('Patient not found.');return;}const receiptNo=payment.receiptNo||payment.id,phone=(patient.phone||'').replace(/\D/g,'');if(!phone){alert('No WhatsApp number saved for this patient.');return;}const international=phone.length===10?'91'+phone:phone;const msg=`Dear ${patient.name},\n\nPlease find your payment receipt ${receiptNo}.\nPayment received: ₹${money(payment.paid)}\n\nThank you. ${brandClinic()}\n${brandDoctor()}`;const blob=buildReceiptBlob(patient,payment);const file=new File([blob],`Receipt_${receiptNo}.pdf`,{type:'application/pdf'});try{if(navigator.share&&(!navigator.canShare||navigator.canShare({files:[file]}))){await navigator.share({title:`Receipt ${receiptNo}`,text:msg,files:[file]});return;}}catch(err){if(err&&err.name==='AbortError')return;}const url=URL.createObjectURL(blob),a=document.createElement('a');a.href=url;a.download=`Receipt_${receiptNo}.pdf`;document.body.appendChild(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(url),3000);window.open('https://wa.me/'+international+'?text='+encodeURIComponent(msg),'_blank');};
  const originalRenderDash=renderDash;renderDash=function(){originalRenderDash();refreshBranding();};
  const originalShow=show;show=function(id,btn){originalShow(id,btn);if(id==='settings'){setTimeout(addBrandSettings,0);setTimeout(refreshBranding,0);}};
  addBrandSettings();refreshBranding();
})();`

func buildHTML() string {
    png := "data:image/png;base64," + base64.StdEncoding.EncodeToString(appIconPNG)
    jpg := base64.StdEncoding.EncodeToString(appIconJPEG)
    js := strings.ReplaceAll(brandingJS, "__APP_LOGO_DATA__", fmt.Sprintf("%q", png))
    js = strings.ReplaceAll(js, "__APP_LOGO_JPEG__", fmt.Sprintf("%q", jpg))
    return strings.Replace(string(indexHTML), "</body>", "<script>"+js+"</script></body>", 1)
}

func main() {
    db := dataFile(); ensureDB(db); pageHTML := buildHTML()
    mux := http.NewServeMux()
    mux.HandleFunc("/api/db", func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Access-Control-Allow-Origin", "*"); w.Header().Set("Content-Type", "application/json"); switch r.Method { case http.MethodGet: _,_=w.Write(readDB(db)); case http.MethodPut: if err:=writeDB(db,r.Body);err!=nil{http.Error(w,`{"ok":false}`,http.StatusBadRequest);return};_,_=w.Write([]byte(`{"ok":true}`)); default:http.Error(w,"method not allowed",http.StatusMethodNotAllowed)}})
    mux.HandleFunc("/", func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"&&r.URL.Path!="/index.html"{http.NotFound(w,r);return};w.Header().Set("Content-Type","text/html; charset=utf-8");_,_=io.WriteString(w,pageHTML)})
    ln,err:=net.Listen("tcp","127.0.0.1:0");if err!=nil{panic(err)};srv:=&http.Server{Handler:mux,ReadHeaderTimeout:5*time.Second};url:=fmt.Sprintf("http://%s/",ln.Addr().String());go func(){_=srv.Serve(ln)}()
    w:=webview2.NewWithOptions(webview2.WebViewOptions{Debug:false,AutoFocus:true,WindowOptions:webview2.WindowOptions{Title:"Amit Patel Dental Software",Width:1440,Height:900,IconId:1,Center:true}});if w==nil{panic("failed to create WebView2 window")};defer w.Destroy();w.SetSize(1440,900,webview2.HintMax);w.Navigate(url);w.Run()
}
