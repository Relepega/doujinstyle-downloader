package v2

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(v)
}

func (ws *webserver) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found, try something else..."))
}

func (ws *webserver) handleBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad request, try something else..."))
}

func (ws *webserver) handleIndexRoute(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ws.templates.ExecuteWithWriter(w, "index", ws.q.GetUIData())
}

func (ws *webserver) handleRestartServer(w http.ResponseWriter, r *http.Request) {
	self, err := os.Executable()
	if err != nil {
		return
	}

	args := os.Args
	env := os.Environ()

	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env

		err := cmd.Run()
		if err == nil {
			os.Exit(0)
		}

	}

	err = syscall.Exec(self, args, env)
	if err != nil {
		return
	}
}
