package v2

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl"
	"github.com/relepega/doujinstyle-downloader/internal/dsdl/db/states"
	"github.com/relepega/doujinstyle-downloader/internal/task"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/sse"
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
	log.Println("Webserver: ", err)

	e := sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()
	ws.msgChan <- e

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, err.Error())
}

func (ws *Webserver) handleTaskAdd(w http.ResponseWriter, r *http.Request) {
	service := strings.TrimSpace(r.FormValue("Service"))
	slugs := r.FormValue("Slugs")

	log.Printf(
		"WebServer: New request: HandleTaskAdd: service: \"%v\" slugs: [%v]\n",
		service,
		slugs,
	)

	engine, _ := ws.UserData.(*dsdl.DSDL)

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
		err := engine.Tq.EnqueueFromValueWithComparator(
			newTask,
			func(item, target any) bool {
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
	taskIDs := r.FormValue("IDs")
	mode := strings.TrimSpace(r.FormValue("Mode"))
	// fmt.Println("mode", mode)

	if !isValidMode(mode, validRemoveModes) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Not a valid mode")
		return
	}

	engine, _ := ws.UserData.(*dsdl.DSDL)

	var happenedErrors []string

	if mode == "single" || mode == "multiple" {
		delimiter := "|"

		if taskIDs == "" || taskIDs == delimiter {
			ws.handleError(w, fmt.Errorf("At least one Album ID is required"))
			return
		}

		idList := strings.SplitSeq(taskIDs, delimiter)

		for id := range idList {
			node, err := engine.Tq.FindWithComparator(id, func(item, target any) bool {
				t := item.(*task.Task)
				id := target.(string)

				return t.Id == id
			})
			if err != nil {
				happenedErrors = append(happenedErrors, err.Error())
				continue
			}

			t := node.(*task.Task)
			t.DownloadState = states.TASK_STATE_QUEUED
			t.Err = nil

			err = engine.Tq.ResetTaskState(node)
			if err != nil {
				happenedErrors = append(happenedErrors, err.Error())
				continue
			}

			tmpl, err := ws.templates.Execute("task", node)
			if err != nil {
				happenedErrors = append(happenedErrors, err.Error())
				continue
			}

			uievt := sse.NewUIEventBuilder().
				Event(sse.UIEvent_ReplaceNode).
				TargetNodeID(id).
				ReceiverNodeSelector("#queued").
				Content(tmpl).
				Position(sse.UIRenderPos_BeforeEnd).
				Build()

			ws.msgChan <- sse.NewSSEBuilder().Event("replace-node").Data(uievt).Build()
		}

		if len(happenedErrors) != 0 {
			ws.handleError(w, fmt.Errorf("%+v", happenedErrors))
			return
		}

		goto retNoErr
	}

	switch mode {
	case "failed":
		nodes, err := engine.Tq.FindWithProgressState(states.TASK_STATE_COMPLETED)
		if err != nil {
			ws.handleError(w, err)
		}

		for _, node := range nodes {
			t := node.(*task.Task)

			if t.Err == nil {
				continue
			}

			t.DownloadState = states.TASK_STATE_QUEUED
			t.Err = nil

			err := engine.Tq.ResetTaskState(node)
			if err != nil {
				happenedErrors = append(happenedErrors, err.Error())
				continue
			}

			tmpl, err := ws.templates.Execute("task", t)
			if err != nil {
				happenedErrors = append(happenedErrors, err.Error())
				continue
			}

			uievt := sse.NewUIEventBuilder().
				Event(sse.UIEvent_ReplaceNode).
				TargetNodeID(t.Id).
				ReceiverNodeSelector("#queued").
				Content(tmpl).
				Position(sse.UIRenderPos_BeforeEnd).
				Build()

			ws.msgChan <- sse.NewSSEBuilder().Event("replace-node").Data(uievt).Build()
		}

		if len(happenedErrors) != 0 {
			ws.handleError(w, fmt.Errorf("%+v", happenedErrors))
			return
		}
	}

retNoErr:
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

	tasks, err := engine.Tq.GetDatabase().GetAll()
	if err != nil {
		ws.handleInternalServerError(w, r, err.Error())
	}

	sendMultiUpdate := func(division string) {
		t, _ := ws.templates.Execute(division+"_tasks", tasks)

		uievt := sse.NewUIEventBuilder().
			Event(sse.UIEvent_ReplaceNodeContent).
			ReceiverNodeSelector(division).
			Content(t).
			Position(sse.UIRenderPos_AfterBegin).
			Build()

		ws.msgChan <- sse.NewSSEBuilder().Event("update-node-content").Data(uievt).Build()
	}

	var happenedErrors []string

	if mode == "single" || mode == "multiple" {
		delimiter := "|"

		if taskIDs == "" || taskIDs == delimiter {
			ws.handleError(w, fmt.Errorf("At least one Album ID is required"))
			return
		}

		idList := strings.SplitSeq(taskIDs, delimiter)

		for id := range idList {
			err := engine.Tq.RemoveWithComparator(id, func(item, target any) bool {
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
			return
		}

		goto retNoErr
	}

	switch mode {
	case "queued":
		_, err := engine.Tq.RemoveFromState(states.TASK_STATE_QUEUED)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate("queued")

	case "completed":
		_, err := engine.Tq.RemoveFromState(states.TASK_STATE_COMPLETED)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate("ended")

	case "failed":
		err := engine.Tq.RemoveWithComparator(
			states.TASK_STATE_COMPLETED,
			func(item, target any) bool {
				t := item.(*task.Task)
				state := target.(int)

				return t.DownloadState == state && t.Err != nil
			},
		)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate("ended")

	case "succeeded":
		err := engine.Tq.RemoveWithComparator(
			states.TASK_STATE_COMPLETED,
			func(item, target any) bool {
				t := item.(*task.Task)
				state := target.(int)

				return t.DownloadState == state && t.Err == nil
			},
		)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		sendMultiUpdate("ended")

	}

retNoErr:
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w)
}
