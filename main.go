package main

import (
    _ "embed"
    "encoding/json"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
    "sync"
    "time"
)

// Exact master HTML is embedded unchanged. The durable save layer is injected
// only for the Windows application build.
//go:embed Dental_Patient_Management_Software.html
var indexHTML []byte
//go:embed save-fix.js
var saveFix []byte

var dbMu sync.Mutex

func dataFile() string {
    base, err := os.UserConfigDir()
    if err != nil { base, _ = os.UserHomeDir() }
    dir := filepath.Join(base, "Amit Patel Dental Software V2")
    _ = os.MkdirAll(dir, 0700)
    return filepath.Join(dir, "patient-data.json")
}

func ensureDB(path string) {
    dbMu.Lock(); defer dbMu.Unlock()
    if _, err := os.Stat(path); err == nil { return }
    initial := `{"patients":[],"visits":[],"payments":[],"appointments":[],"suggestions":{"complaint":[],"diagnosis":[],"treatment":[],"instructions":[]},"medicines":[],"prescriptionTemplates":[],"settings":{}}`
    _ = os.WriteFile(path, []byte(initial), 0600)
}

func readDB(path string) []byte {
    dbMu.Lock(); defer dbMu.Unlock()
    b, err := os.ReadFile(path)
    if err != nil { return []byte(`{"patients":[],"visits":[],"payments":[],"appointments":[],"suggestions":{"complaint":[],"diagnosis":[],"treatment":[],"instructions":[]},"medicines":[],"prescriptionTemplates":[],"settings":{}}`) }
    return b
}

func writeDB(path string, r io.Reader) error {
    var value any
    if err := json.NewDecoder(r).Decode(&value); err != nil { return err }
    data, err := json.MarshalIndent(value, "", "  ")
    if err != nil { return err }
    dbMu.Lock(); defer dbMu.Unlock()
    dir := filepath.Dir(path); _ = os.MkdirAll(dir, 0700)
    tmp, err := os.CreateTemp(dir, "patient-data-*.tmp")
    if err != nil { return err }
    tmpName := tmp.Name(); defer os.Remove(tmpName)
    if err := tmp.Chmod(0600); err != nil { _ = tmp.Close(); return err }
    if _, err := tmp.Write(data); err != nil { _ = tmp.Close(); return err }
    if err := tmp.Sync(); err != nil { _ = tmp.Close(); return err }
    if err := tmp.Close(); err != nil { return err }
    return os.Rename(tmpName, path)
}

func launch(url string) {
    if runtime.GOOS != "windows" { return }
    for _, browser := range []string{"msedge.exe", "chrome.exe"} {
        if _, err := exec.LookPath(browser); err == nil {
            _ = exec.Command(browser, "--app="+url, "--new-window", "--disable-features=TranslateUI").Start(); return
        }
    }
    _ = exec.Command("cmd", "/c", "start", "", url).Start()
}

func main() {
    db := dataFile(); ensureDB(db)
    mux := http.NewServeMux()
    mux.HandleFunc("/api/db", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        switch r.Method {
        case http.MethodGet: _, _ = w.Write(readDB(db))
        case http.MethodPut:
            if err := writeDB(db, r.Body); err != nil { http.Error(w, `{"ok":false}`, 400); return }
            _, _ = w.Write([]byte(`{"ok":true}`))
        default: http.Error(w, "method not allowed", 405)
        }
    })
    mux.HandleFunc("/api/db-beacon", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost { _ = writeDB(db, r.Body) }
        w.WriteHeader(http.StatusNoContent)
    })
    mux.HandleFunc("/save-fix.js", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/javascript; charset=utf-8"); _, _ = w.Write(saveFix)
    })
    page := strings.Replace(string(indexHTML), "</body>", `<script src="/save-fix.js"></script></body>`, 1)
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" && r.URL.Path != "/index.html" { http.NotFound(w, r); return }
        w.Header().Set("Content-Type", "text/html; charset=utf-8"); _, _ = io.WriteString(w, page)
    })
    ln, err := net.Listen("tcp", "127.0.0.1:0"); if err != nil { panic(err) }
    srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
    go func() { _ = srv.Serve(ln) }()
    launch(fmt.Sprintf("http://%s/", ln.Addr()))
    select {}
}
