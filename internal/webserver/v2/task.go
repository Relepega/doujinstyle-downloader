package v2

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/task"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/SSEEvents"
)

var (
	validRemoveModes = []string{"single", "multiple", "queued", "failed", "succeeded"}
	validUpdateModes = []string{"single", "multiple", "failed"}
)

type TaskEvent struct {
	AlbumID     string `json:"AlbumID"`
	GroupAction string `json:"GroupAction"`
}

func isValidMode(m string, ms []string) bool {
	isValid := false
	for _, v := range ms {
		if v == m {
			isValid = true
		}
	}

	return isValid
}

func (ws *Webserver) handleError(w http.ResponseWriter, err error) {
	log.Println("error: ", err)

	ws.msgChan <- SSEEvents.NewSSEMessageWithError(err)

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, err.Error())
}

func (ws *Webserver) handleTaskAdd(w http.ResponseWriter, r *http.Request) {
	engine, _ := ws.UserData.(*dsdl.DSDL)

	slugs := r.FormValue("Slugs")
	service := strings.TrimSpace(r.FormValue("Service"))

	delimiter := "|"

	if slugs == "" || slugs == delimiter {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "At least one Album Slug is required")
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

		// set values to struct fields
		newTask := &task.Task{
			AggregatorName: service,
			AggregatorSlug: slug,
			DisplayName:    slug,
		}
		newTask.AggregatorName = service

		if strings.HasPrefix(slug, "http") {
			newTask.AggregatorPageURL = slug
		}

		// add task to engine
		err := engine.Tq.AddNodeFromValueWithComparator(
			newTask,
			func(item, target interface{}) bool {
				toCompare := item.(*task.Task)

				return toCompare.AggregatorSlug == newTask.AggregatorSlug
			},
		)
		if err != nil {
			happenedErrors = append(happenedErrors, err.Error())
			continue
		}

		// render template
		t, err := ws.templates.Execute("task", newTask)
		if err != nil {
			ws.msgChan <- SSEEvents.NewSSEMessageWithError(err)

			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err.Error())

			return
		}
		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("new-task", t)
	}

	if len(happenedErrors) != 0 {
		ws.msgChan <- SSEEvents.NewSSEMessageWithError(fmt.Errorf("%+v", happenedErrors))
	}

	// :)
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

	err = engine.Tq.ResetTaskState(taskVal)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err.Error())
		return
	}

	t, err := ws.templates.Execute("task", taskVal)
	if err != nil {
		ws.handleError(w, err)
		return
	}

	renderEvt := SSEEvents.NewUIRenderEvent(
		SSEEvents.ReplaceNode,
		slug,
		"#queued",
		t,
		SSEEvents.AfterBegin,
	)
	val, err := renderEvt.String()
	if err != nil {
		ws.handleError(w, err)
		return
	}
	ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node", val)

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

	delimiter := "|"

	if slugs == "" || slugs == delimiter {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "At least one Album Slug is required")
		return
	}

	slugList := strings.Split(slugs, delimiter)

	engine, _ := ws.UserData.(*dsdl.DSDL)

	for _, s := range slugList {
		if s == "" {
			continue
		}

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

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("remove-task", s)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w)
}
