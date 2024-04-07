package webserver

import (
	pubsub "github.com/relepega/doujinstyle-downloader-reloaded/internal/pubSub"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/taskQueue"
	"github.com/relepega/doujinstyle-downloader-reloaded/internal/webserver/SSEEvents"
)

type SSEMsgBrokerEvt string

const (
	ActivateTaskEvent   SSEMsgBrokerEvt = "activate-task"
	MarkTaskAsDoneEvent SSEMsgBrokerEvt = "mark-task-as-done"
	ErrorEvent          SSEMsgBrokerEvt = "error"
)

func (ws *webserver) SSEMsgBroker() {
	subscriber := pubsub.GetExistingPublisher().Subscribe()

	for {
		select {
		case msg := <-subscriber:
			switch msg.EvtType {
			case "activate-task":
				t, err := ws.templates.Execute("task", msg.Data)
				if err != nil {
					ws.msgChan <- SSEEvents.NewSSEMessageWithError(err)
					continue
				}

				// fmt.Println("activated task", t)
				nodeId := msg.Data.(*taskQueue.Task).AlbumID
				s, _ := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNode, nodeId, "#active", t, SSEEvents.BeforeEnd).String()

				ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node", s)

			case "mark-task-as-done":
				t, _ := ws.templates.Execute("task", msg.Data)
				// fmt.Println("re-rendered task: ", t)

				nodeId := msg.Data.(*taskQueue.Task).AlbumID
				s, _ := SSEEvents.NewUIRenderEvent(SSEEvents.ReplaceNode, nodeId, "#ended", t, SSEEvents.AfterBegin).String()

				ws.msgChan <- SSEEvents.NewSSEMessageWithEvent("replace-node", s)

			case "error":
				ws.msgChan <- SSEEvents.NewSSEMessageWithError(msg.Data.(error))
			}

		default:
			continue

		}
	}
}