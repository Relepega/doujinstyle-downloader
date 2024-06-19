package tq_eventbroker

import (
	pubsub "github.com/relepega/doujinstyle-downloader/internal/pubSub"
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
				queuePub.Publish(&pubsub.PublishEvent{
					EvtType: "update-task-progress",
					Data:    evt.Data,
				})

			case "update-task-name":
				queuePub.Publish(&pubsub.PublishEvent{
					EvtType: "update-task-name",
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
