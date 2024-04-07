package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/SSEEvents"
)

type TaskEvent struct {
	AlbumID     string `json:"AlbumID"`
	GroupAction string `json:"GroupAction"`
}

func (ws *webserver) handleError(w http.ResponseWriter, err error) {
	log.Println("error: ", err)

	ws.msgChan <- SSEEvents.NewSSEMessageWithError(err)

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, err.Error())
}

func (ws *webserver) handleTaskAdd(w http.ResponseWriter, r *http.Request) {
	albumID := r.FormValue("AlbumID")
	serviceNumberStr := r.FormValue("ServiceNumber")

	if albumID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "AlbumID is required")
		return
	}

	if serviceNumberStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "ServiceNumber is required")
		return
	}

	// tasks = append(tasks, albumID)
	newTask := taskQueue.NewTask(albumID, 0)
	err := ws.q.AddTask(newTask)
	if err != nil {
		ws.msgChan <- SSEEvents.NewSSEMessageWithError(err)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())

		return
	}

	t, err := ws.templates.Execute("task", newTask)
	if err != nil {
		ws.msgChan <- SSEEvents.NewSSEMessageWithError(err)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())

		return
	}

	ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("new-task", t)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, albumID, serviceNumberStr)
}

func (ws *webserver) handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	var data TaskEvent

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	if data.GroupAction == "" && data.AlbumID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "GroupAction or AlbumID is required")
		return
	}

	// fmt.Println(data.AlbumID)

	switch data.GroupAction {
	case "":
		err := ws.q.RemoveTaskFromAlbumID(data.AlbumID)
		if err != nil {
			ws.handleError(w, err)
			return
		}

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("remove-task", data.AlbumID)

	case "clear-queued":
		ws.q.ClearQueuedTasks()

		t, err := ws.templates.Execute("queued_tasks", ws.q.GetUIData())
		if err != nil {
			ws.handleError(w, err)
			return
		}

		renderEvt := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNodeContent, "", "queued", t, SSEEvents.AfterBegin)
		val, err := renderEvt.String()
		if err != nil {
			ws.handleError(w, err)
			return
		}

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node-content", val)

	case "clear-all-completed":
		ws.q.ClearAllCompleted()

		t, err := ws.templates.Execute("ended_tasks", ws.q.GetUIData())
		if err != nil {
			ws.handleError(w, err)
			return
		}

		renderEvt := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNodeContent, "", "ended", t, SSEEvents.AfterBegin)
		val, err := renderEvt.String()
		if err != nil {
			ws.handleError(w, err)
			return
		}

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node-content", val)

	case "clear-fail-completed":
		ws.q.ClearFailedCompleted()

		t, err := ws.templates.Execute("ended_tasks", ws.q.GetUIData())
		if err != nil {
			ws.handleError(w, err)
			return
		}

		renderEvt := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNodeContent, "", "ended", t, SSEEvents.AfterBegin)
		val, err := renderEvt.String()
		if err != nil {
			ws.handleError(w, err)
			return
		}

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node-content", val)

	case "clear-success-completed":
		ws.q.ClearSuccessfullyCompleted()

		t, err := ws.templates.Execute("ended_tasks", ws.q.GetUIData())
		if err != nil {
			ws.handleError(w, err)
			return
		}

		renderEvt := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNodeContent, "", "ended", t, SSEEvents.AfterBegin)
		val, err := renderEvt.String()
		if err != nil {
			ws.handleError(w, err)
			return
		}

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node-content", val)

	case "retry-fail-completed":
		ws.q.ResetFailed()

		t, err := ws.templates.Execute("task_controls", ws.q.GetUIData())
		if err != nil {

			ws.handleError(w, err)
			return
		}

		renderEvt := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNodeContent, "", "tasks-controls", t, SSEEvents.AfterBegin)
		val, err := renderEvt.String()
		if err != nil {
			ws.handleError(w, err)
			return
		}

		ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node-content", val)

	}
}

func (ws *webserver) handleTaskRetry(w http.ResponseWriter, r *http.Request) {
	albumID := r.FormValue("AlbumID")

	if albumID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "AlbumID is required")
		return
	}

	ws.q.ResetTask(albumID)

	task, err := ws.q.GetTask(albumID)
	if err != nil {
		ws.handleError(w, err)
		return
	}

	t, err := ws.templates.Execute("task", task)
	if err != nil {
		ws.handleError(w, err)
		return
	}

	renderEvt := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNode, albumID, "#queued", t, SSEEvents.AfterBegin)
	val, err := renderEvt.String()
	if err != nil {
		ws.handleError(w, err)
		return
	}

	ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node", val)
}