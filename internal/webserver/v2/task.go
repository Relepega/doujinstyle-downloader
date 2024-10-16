package v2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/downloader/services"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue/task"
)

func (ws *webserver) handleTaskAdd(w http.ResponseWriter, r *http.Request) {
	slugs := strings.TrimSpace(r.FormValue("ServiceSlugs"))
	service := r.FormValue("Service")

	delimiter := "|"

	if slugs == "" || slugs == delimiter {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "At least one AlbumID is required")
		return
	}

	if service == "" || !services.IsValidService(service) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Not a valid service")
		return
	}

	slugList := strings.Split(slugs, delimiter)

	for _, slug := range slugList {
		if slug == "" {
			continue
		}

		task := task.NewTaskFromSlug(slug)

		ws.tq.AddNodeFromValue(task)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, slugList, service)
}

func (ws *webserver) handleTaskUpdate(w http.ResponseWriter, r *http.Request) {}

func (ws *webserver) handleTaskRemove(w http.ResponseWriter, r *http.Request) {}
