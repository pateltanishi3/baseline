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
    "sync"
    "time"

    webview2 "github.com/jchv/go-webview2"
)

//go:embed Dental_Patient_Management_Software.html
var indexHTML []byte

//go:embed branding.js
var brandingJS []byte

//go:embed app-icon.svg
var appIconSVG []byte

//go:embed save-fix.js
var saveFixJS []byte

var emptyDB = map[string]any{
    "patients": []any{}, "visits": []any{}, "payments": []any{}, "appointments": []any{},
    "suggestions": map[string]any{"complaint": []any{}, "diagnosis": []any{}, "treatment": []any{}, "instructions": []any{}},
    "medicines": []any{}, "prescriptionTemplates": []any{},
    "settings": map[string]any{"theme": "ocean", "rxWidth": 145, "rxHeight": 190, "rxTop": 0, "rxRight": 0, "rxBottom": 0, "rxLeft": 0, "doctorName": "Dr. Amit J. Patel", "clinicName": "Dr. Patel's Dental Implant Center & Clinic"},
}

var dbMu sync.RWMutex

func dataFile() string {
    base, err := os.UserConfigDir(); if err != nil { base, _ = os.UserHomeDir() }
    dir := filepath.Join(base, "Amit Patel Dental Software"); _ = os.MkdirAll(dir, 0700)
    return filepath.Join(dir, "patient-data.json")
}
func ensureDB(p string) { dbMu.Lock(); defer dbMu.Unlock(); if _, err := os.Stat(p); err == nil { return }; b, _ := json.MarshalIndent(emptyDB, "", "  "); _ = os.WriteFile(p, b, 0600) }
func readDB(p string) []byte { dbMu.RLock(); defer dbMu.RUnlock(); b, err := os.ReadFile(p); if err != nil { b, _ = json.Marshal(emptyDB) }; return b }
func writeDB(p string, r io.Reader) error {
    dbMu.Lock(); defer dbMu.Unlock()
    var v any
    if err := json.NewDecoder(r).Decode(&v); err != nil { return err }
    // Reject a stale snapshot when a newer version is already on disk.
    if incoming, ok := v.(map[string]any); ok {
        incomingMeta, _ := incoming["_meta"].(map[string]any)
        incomingVersion, _ := incomingMeta["version"].(float64)
        if incomingVersion > 0 {
            if current, err := os.ReadFile(p); err == nil {
                var currentObj map[string]any
                if json.Unmarshal(current, &currentObj) == nil {
                    currentMeta, _ := currentObj["_meta"].(map[string]any)
                    currentVersion, _ := currentMeta["version"].(float64)
                    if currentVersion > incomingVersion { return nil }
                }
            }
        }
    }
    b, err := json.MarshalIndent(v, "", "  "); if err != nil { return err }
    dir := filepath.Dir(p); _ = os.MkdirAll(dir, 0700)
    tmp, err := os.CreateTemp(dir, "patient-data-*.tmp"); if err != nil { return err }
    tmpName := tmp.Name()
    defer os.Remove(tmpName)
    if err := tmp.Chmod(0600); err != nil { tmp.Close(); return err }
    if _, err := tmp.Write(b); err != nil { tmp.Close(); return err }
    if err := tmp.Sync(); err != nil { tmp.Close(); return err }
    if err := tmp.Close(); err != nil { return err }
    return os.Rename(tmpName, p)
}

func buildHTML() string {
    svg := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(appIconSVG)
    js := strings.ReplaceAll(string(brandingJS), "__APP_LOGO_DATA__", fmt.Sprintf("%q", svg))
    saveFix := string(saveFixJS)
    html := strings.Replace(string(indexHTML), "</body>", "<script>"+js+"</script><script>"+saveFix+"</script></body>", 1)
    return html
}

func main() {
    db := dataFile(); ensureDB(db); pageHTML := buildHTML()
    mux := http.NewServeMux()
    mux.HandleFunc("/api/db", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*"); w.Header().Set("Content-Type", "application/json")
        switch r.Method {
        case http.MethodGet:
            _, _ = w.Write(readDB(db))
        case http.MethodPut:
            if err := writeDB(db, r.Body); err != nil { http.Error(w, `{"ok":false}`, http.StatusBadRequest); return }
            _, _ = w.Write([]byte(`{"ok":true}`))
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    })
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" && r.URL.Path != "/index.html" { http.NotFound(w, r); return }; w.Header().Set("Content-Type", "text/html; charset=utf-8"); _, _ = io.WriteString(w, pageHTML) })
    ln, err := net.Listen("tcp", "127.0.0.1:0"); if err != nil { panic(err) }
    srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}; url := fmt.Sprintf("http://%s/", ln.Addr().String()); go func() { _ = srv.Serve(ln) }()
    w := webview2.NewWithOptions(webview2.WebViewOptions{Debug:false, AutoFocus:true, WindowOptions:webview2.WindowOptions{Title:"Amit Patel Dental Software", Width:1440, Height:900, IconId:1, Center:true}})
    if w == nil { panic("failed to create WebView2 window") }; defer w.Destroy(); w.SetSize(1440, 900, webview2.HintMax); w.Navigate(url); w.Run()
}
