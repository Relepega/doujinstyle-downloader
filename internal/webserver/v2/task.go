package v2

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/task"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/v2/sse"
)

var (
	validRemoveModes = []string{"single", "multiple", "queued", "completed", "failed", "succeeded"}
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

	e := sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()
	ws.msgChan <- e

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
		newTask := task.NewTask(slug)
		newTask.Aggregator = service

		if strings.HasPrefix(slug, "http") {
			newTask.AggregatorPageURL = slug
		}

		// add task to engine
		err := engine.Tq.AddNodeFromValueWithComparator(
			newTask,
			func(item, target interface{}) bool {
				toCompare := item.(*task.Task)

				return toCompare.Slug == newTask.Slug
			},
		)
		if err != nil {
			happenedErrors = append(happenedErrors, err.Error())
			continue
		}

		// render template
		t, err := ws.templates.Execute("task", newTask)
		if err != nil {
			ws.msgChan <- sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()

			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err.Error())

			return
		}
		ws.msgChan <- sse.NewSSEBuilder().Event("new-task").Data(appUtils.CleanString(t)).Build()
	}

	if len(happenedErrors) != 0 {
		ws.msgChan <- sse.NewSSEBuilder().Event("error").Data(fmt.Errorf("%+v", happenedErrors).Error()).Build()
	}

	// :)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, slugList, service, happenedErrors)
}

func (ws *Webserver) handleTaskUpdateState(w http.ResponseWriter, r *http.Request) {
	engine, _ := ws.UserData.(*dsdl.DSDL)

	nodeID := r.FormValue("Id")

	if nodeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "A node ID is required")
		return
	}

	taskVal, err := engine.Tq.GetNodeWithComparator(nodeID, func(item, target interface{}) bool {
		i := item.(*task.Task)
		t := target.(string)

		return i.Slug == t
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

	uievt := sse.NewUIEventBuilder().
		Event(sse.UIEvent_ReplaceNode).
		TargetNodeID(nodeID).
		ReceiverNodeSelector("#queued").
		Content(appUtils.CleanString(t)).
		Position(sse.UIRenderPos_AfterBegin).
		Build()

	ws.msgChan <- sse.NewSSEBuilder().Event("replace-node").Data(uievt).Build()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w)
}

func (ws *Webserver) handleTaskRemove(w http.ResponseWriter, r *http.Request) {
	taskIDs := r.FormValue("IDs")
	mode := strings.TrimSpace(r.FormValue("Mode"))
	// fmt.Println("mode", mode)

	if !isValidMode(mode, validRemoveModes) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Not a valid mode")
		return
	}

	engine, _ := ws.UserData.(*dsdl.DSDL)

	sendMultiUpdate := func() {
		t, _ := ws.templates.Execute("ended_tasks", engine.Tq.GetTracker().GetAll())

		uievt := sse.NewUIEventBuilder().
			Event(sse.UIEvent_ReplaceNodeContent).
			ReceiverNodeSelector("ended").
			Content(t).
			Position(sse.UIRenderPos_AfterBegin).
			Build()

		ws.msgChan <- sse.NewSSEBuilder().Event("update-node-content").Data(uievt).Build()
	}

	// var event string
	var happenedErrors []string

	if mode == "single" || mode == "multiple" {
		delimiter := "|"

		if taskIDs == "" || taskIDs == delimiter {
			ws.handleError(w, fmt.Errorf("At least one Album ID is required"))
			return
		}

		idList := strings.Split(taskIDs, delimiter)

		for _, id := range idList {
			err := engine.Tq.RemoveNodeWithComparator(id, func(item, target interface{}) bool {
				id := target.(string)
				t := item.(*task.Task)

				return t.Id == id
			})
			if err != nil {
				happenedErrors = append(happenedErrors, err.Error())
			} else {
				ws.msgChan <- sse.NewSSEBuilder().Event("remove-node").Data(id).Build()
			}
		}

		if len(happenedErrors) != 0 {
			ws.handleError(w, fmt.Errorf("%+v", happenedErrors))
		}

		goto retNoErr
	}

	switch mode {
	case "queued":
		_, err := engine.Tq.RemoveFromState(dsdl.TASK_STATE_QUEUED)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		goto retNoErr

	case "completed":
		_, err := engine.Tq.RemoveFromState(dsdl.TASK_STATE_COMPLETED)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate()

		goto retNoErr

	case "failed":
		err := engine.Tq.RemoveNodeWithComparator(
			dsdl.TASK_STATE_COMPLETED,
			func(item, target interface{}) bool {
				t := item.(*task.Task)
				state := target.(int)

				return t.DownloadState == state && t.Err == nil
			},
		)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate()

		goto retNoErr

	case "succeeded":
		err := engine.Tq.RemoveNodeWithComparator(
			dsdl.TASK_STATE_COMPLETED,
			func(item, target interface{}) bool {
				t := item.(*task.Task)
				state := target.(int)

				return t.DownloadState == state && t.Err != nil
			},
		)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate()

		goto retNoErr

	}

retNoErr:
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w)
}
