package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// systemDirs maps the ESR registry name to the directory inside mbaigo-systems/.
var systemDirs = map[string]string{
	"serviceregistrar": "esr",
	"orchestrator":     "orchestrator",
	"modboss":          "modboss",
	"telegrapher":      "telegrapher",
	"opcuac":           "uaclient",
}

var (
	mu           sync.Mutex
	processes    = map[string]*exec.Cmd{}
	servicesRoot string
)

func jsonResp(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	relDir, ok := systemDirs[name]
	if !ok {
		jsonResp(w, http.StatusNotFound, "unknown system: "+name)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, running := processes[name]; running {
		jsonResp(w, http.StatusConflict, name+" is already running")
		return
	}

	dir := filepath.Join(servicesRoot, relDir)
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		jsonResp(w, http.StatusInternalServerError, fmt.Sprintf("start failed: %v", err))
		return
	}

	processes[name] = cmd
	log.Printf("started %s (pid %d) from %s", name, cmd.Process.Pid, dir)

	// Remove from map when the process exits on its own.
	go func() {
		_ = cmd.Wait()
		mu.Lock()
		delete(processes, name)
		mu.Unlock()
		log.Printf("%s exited", name)
	}()

	jsonResp(w, http.StatusOK, "started")
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if _, ok := systemDirs[name]; !ok {
		jsonResp(w, http.StatusNotFound, "unknown system: "+name)
		return
	}
	mu.Lock()
	_, running := processes[name]
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"running": running})
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if _, ok := systemDirs[name]; !ok {
		jsonResp(w, http.StatusNotFound, "unknown system: "+name)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	cmd, running := processes[name]
	if !running {
		jsonResp(w, http.StatusNotFound, name+" was not started by this backend")
		return
	}

	if err := killProcess(cmd); err != nil {
		jsonResp(w, http.StatusInternalServerError, fmt.Sprintf("kill failed: %v", err))
		return
	}

	delete(processes, name)
	log.Printf("stopped %s", name)
	jsonResp(w, http.StatusOK, "stopped")
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot determine working directory: %v", err)
	}
	// Expected to run from mbaigo-systems/backend/ — services live in the parent dir.
	servicesRoot = filepath.Join(wd, "..")
	log.Printf("services root: %s", servicesRoot)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/systems/{name}/start", handleStart)
	mux.HandleFunc("POST /api/systems/{name}/stop", handleStop)
	mux.HandleFunc("GET /api/systems/{name}/status", handleStatus)

	log.Println("dashboard backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
