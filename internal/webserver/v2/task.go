package v2

import (
	"fmt"
	"net/http"
	"strings"

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
	engine, _ := ws.UserData.(*dsdl.DSDL)

	slugs := r.FormValue("Slugs")
	service := strings.TrimSpace(r.FormValue("Service"))

	delimiter := "|"

	if slugs == "" || slugs == delimiter {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "At least one AlbumID is required")
		return
	}

	if service == "" || !engine.IsValidAggregator(service) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Not a valid service")
		return
	}

	slugList := strings.Split(slugs, delimiter)

	var happenedErrors []string

	for _, slug := range slugList {
		if slug == "" {
			continue
		}

		newTask := task.NewTaskFromSlug(slug)

		err := engine.Tq.AddNodeFromValueWithComparator(
			newTask,
			func(item, target interface{}) bool {
				toCompare := item.(*task.Task)

				return toCompare.AggregatorSlug == newTask.AggregatorSlug
			},
		)
		if err != nil {
			happenedErrors = append(happenedErrors, err.Error())
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, slugList, service, happenedErrors)
}

func (ws *Webserver) handleTaskUpdateState(w http.ResponseWriter, r *http.Request) {
	engine, _ := ws.UserData.(*dsdl.DSDL)

	slug := r.FormValue("Slug")

	if slug == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "A task slug is required")
		return
	}

	taskVal, err := engine.Tq.GetNodeWithComparator(slug, func(item, target interface{}) bool {
		i := item.(*task.Task)
		t := target.(string)

		return i.AggregatorSlug == t
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Slug not found")
		return
	}

	engine.Tq.ResetTaskState(taskVal)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Slug not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w)
}

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

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w)
}
