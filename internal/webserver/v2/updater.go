package v2

import (
	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
	"github.com/relepega/doujinstyle-downloader/internal/task"
	"github.com/relepega/doujinstyle-downloader/internal/webserver/v2/sse"
)

func (ws *Webserver) parseAndSendUpdates() {
	publisher, err := pubsub.GetGlobalPublisher("task-updater")
	if err != nil {
		publisher = pubsub.NewGlobalPublisher("task-updater")
	}

	subscriber := publisher.Subscribe()

	for {
		select {
		case <-ws.ctx.Done():
			return

		case msg := <-subscriber:
			switch msg.EvtType {
			case "new-task":
				t, err := ws.templates.Execute("task", msg.Data)
				if err != nil {
					e := sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()
					ws.msgChan <- e

					continue
				}

				e := sse.NewSSEBuilder().
					Event("new-task").
					Data(t).
					Build()

				ws.msgChan <- e

			case "activate-task":
				t, err := ws.templates.Execute("task", msg.Data)
				if err != nil {
					e := sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()
					ws.msgChan <- e

					continue
				}

				nodeId := msg.Data.(*task.Task).ID()

				uievt := sse.NewUIEventBuilder().
					Event(sse.UIEvent_ReplaceNode).
					TargetNodeID(nodeId).
					ReceiverNodeSelector("#active").
					Content(appUtils.CleanString(t)).
					Position(sse.UIRenderPos_BeforeEnd).
					Build()

				e := sse.NewSSEBuilder().
					Event("replace-node").
					Data(uievt).
					Build()

				ws.msgChan <- e

			case "mark-task-as-done":
				t, err := ws.templates.Execute("task", msg.Data)
				if err != nil {
					e := sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()
					ws.msgChan <- e
					continue
				}

				nodeId := msg.Data.(*task.Task).ID()

				uievt := sse.NewUIEventBuilder().
					Event(sse.UIEvent_ReplaceNode).
					TargetNodeID(nodeId).
					ReceiverNodeSelector("#ended").
					Content(t).
					Position(sse.UIRenderPos_AfterBegin).
					Build()

				e := sse.NewSSEBuilder().
					Event("replace-node").
					Data(uievt).
					Build()

				ws.msgChan <- e

			case "update-node-content":
				t, err := ws.templates.Execute("task-content", msg.Data)
				if err != nil {
					e := sse.NewSSEBuilder().Event("error").Data(err.Error()).Build()
					ws.msgChan <- e
					continue
				}

				nodeId := msg.Data.(*task.Task).ID()

				t = appUtils.CleanString(t)

				uievt := sse.NewUIEventBuilder().
					Event(sse.UIEvent_ReplaceNodeContent).
					TargetNodeID(nodeId).
					ReceiverNodeSelector(nodeId).
					Content(t).
					Position(sse.UIRenderPos_AfterBegin).
					Build()

				e := sse.NewSSEBuilder().
					Event("update-node-content").
					Data(uievt).
					Build()

				ws.msgChan <- e

			case "error":
				e := sse.NewSSEBuilder().Event("error").Data(msg.Data.(error).Error()).Build()
				ws.msgChan <- e

			}
		}
	}
}
