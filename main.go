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
    "time"
)

//go:embed index.html
var indexHTML []byte

var emptyDB = map[string]any{
    "patients": []any{}, "visits": []any{}, "payments": []any{}, "appointments": []any{},
    "suggestions": map[string]any{"complaint": []any{}, "diagnosis": []any{}, "treatment": []any{}, "instructions": []any{}},
    "medicines": []any{}, "prescriptionTemplates": []any{},
    "settings": map[string]any{"theme": "ocean", "rxWidth": 145, "rxHeight": 190, "rxTop": 0, "rxRight": 0, "rxBottom": 0, "rxLeft": 0},
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
    b, err := os.ReadFile(p); if err != nil { b, _ = json.Marshal(emptyDB) }
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
func openApp(url string) {
    if runtime.GOOS != "windows" { return }
    for _, browser := range []string{"msedge.exe", "chrome.exe"} {
        if _, err := exec.LookPath(browser); err == nil {
            _ = exec.Command(browser, "--app="+url, "--new-window").Start(); return
        }
    }
    _ = exec.Command("cmd", "/c", "start", "", url).Start()
}
func main() {
    db := dataFile(); ensureDB(db)
    mux := http.NewServeMux()
    mux.HandleFunc("/api/db", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Content-Type", "application/json")
        switch r.Method {
        case http.MethodGet: _, _ = w.Write(readDB(db))
        case http.MethodPut:
            if err := writeDB(db, r.Body); err != nil { http.Error(w, `{"ok":false}`, 400); return }
            _, _ = w.Write([]byte(`{"ok":true}`))
        default: http.Error(w, "method not allowed", 405)
        }
    })
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" && r.URL.Path != "/index.html" { http.NotFound(w, r); return }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        _, _ = w.Write(indexHTML)
    })
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { panic(err) }
    srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
    url := fmt.Sprintf("http://%s/", ln.Addr().String())
    go func(){ _ = srv.Serve(ln) }()
    openApp(url)
    select {}
}
