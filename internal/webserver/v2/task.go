package v2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/downloader/services"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue/task"
)

var (
	validRemoveModes = []string{"single", "multiple", "queued", "failed", "succeeded"}
	validUpdateModes = []string{"single", "multiple", "failed"}
)

func isValidMode(m string, ms []string) bool {
	isValid := false
	for _, v := range ms {
		if v == m {
			isValid = true
		}
	}

	return isValid
}

func (ws *Webserver) handleTaskAdd(w http.ResponseWriter, r *http.Request) {
	slugs := r.FormValue("Slugs")
	service := strings.TrimSpace(r.FormValue("Service"))

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

	engine, _ := ws.UserData.(*dsdl.DSDL)

	for _, slug := range slugList {
		if slug == "" {
			continue
		}

		task := task.NewTaskFromSlug(slug)

		engine.Tq.AddNodeFromValue(task)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, slugList, service)
}

func (ws *Webserver) handleTaskUpdate(w http.ResponseWriter, r *http.Request) {}

func (ws *Webserver) handleTaskRemove(w http.ResponseWriter, r *http.Request) {
	slugs := r.FormValue("Slugs")
	mode := strings.TrimSpace(r.FormValue("Mode"))

	if !isValidMode(mode, validRemoveModes) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Not a valid mode")
		return
	}

	engine, _ := ws.UserData.(*dsdl.DSDL)

	for _, s := range slugs {
		engine.Tq.RemoveNodeWithComparator(s, func(int_v, user_v interface{}) bool {
			task, ok := int_v.(task.Task)
			if !ok {
				return false
			}

			slug, ok := user_v.(string)
			if !ok {
				return false
			}

			return task.AggregatorSlug == slug
		})
	}
}
