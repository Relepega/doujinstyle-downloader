package tq_eventbroker

import (
	"fmt"

	pubsub "github.com/relepega/doujinstyle-downloader-reloaded/internal/pubSub"
)

type UpdateTaskProgress struct {
	Id       string
	Progress int8
}

func QueueEvtBroker() {
	queueSub := pubsub.NewGlobalPublisher("queue-broker")
	subscriber := queueSub.Subscribe()

	queuePub, _ := pubsub.GetGlobalPublisher("queue")

	for {
		select {
		case evt := <-subscriber:
			switch evt.EvtType {
			case "update-task-progress":
				fmt.Printf("%+v\n", evt)
				queuePub.Publish(&pubsub.PublishEvent{
					EvtType: "update-task-progress",
					Data:    evt.Data,
				})

			default:
				continue

			}

		default:
			continue

		}
	}
}
