package sse

import (
	"encoding/json"
	"fmt"
	"log"
)

// Actual SSE builder

type SSEBuilder struct {
	event string
	data  string
}

func NewSSEBuilder() *SSEBuilder {
	return &SSEBuilder{
		event: "message",
	}
}

func (sb *SSEBuilder) Event(event string) *SSEBuilder {
	sb.event = event

	return sb
}

func (sb *SSEBuilder) Data(data string) *SSEBuilder {
	sb.data = data

	return sb
}

func (sb *SSEBuilder) BuildWithError() (string, error) {
	if sb.event == "" {
		sb.event = "message"
	}

	if sb.data == "" {
		return "", fmt.Errorf("Data field cannot be empty")
	}

	return fmt.Sprintf("event: %s\ndata: %s\n\n", sb.event, sb.data), nil
}

func (sb *SSEBuilder) Build() string {
	if sb.event == "" {
		sb.event = "message"
	}

	if sb.data == "" {
		log.Fatalln("Data field cannot be empty")
	}

	return fmt.Sprintf("event: %s\ndata: %s\n\n", sb.event, sb.data)
}

// UI Event builder

type UIEvent string

const (
	UIEvent_NewNode            UIEvent = "new-node"
	UIEvent_ReplaceNode        UIEvent = "replace-node"
	UIEvent_ReplaceNodeContent UIEvent = "replace-node-content"
	UIEvent_RemoveNode         UIEvent = "remove-node"
)

type UIRenderPosition string

const (
	UIRenderPos_BeforeBegin UIRenderPosition = "beforebegin"
	UIRenderPos_AfterBegin  UIRenderPosition = "afterbegin"
	UIRenderPos_BeforeEnd   UIRenderPosition = "beforeend"
	UIRenderPos_AfterEnd    UIRenderPosition = "afterend"
	UIRenderPos_Null        UIRenderPosition = ""
)

type UIEventBuilder struct {
	event                              UIEvent
	targetNodeID, receiverNodeSelector string
	newContent                         any
	position                           UIRenderPosition
}

func NewUIEventBuilder() *UIEventBuilder {
	return &UIEventBuilder{}
}

func (ueb *UIEventBuilder) Event(evt UIEvent) *UIEventBuilder {
	ueb.event = evt

	return ueb
}

func (ueb *UIEventBuilder) TargetNodeID(target string) *UIEventBuilder {
	ueb.targetNodeID = target

	return ueb
}

func (ueb *UIEventBuilder) ReceiverNodeSelector(receiver string) *UIEventBuilder {
	ueb.receiverNodeSelector = receiver

	return ueb
}

func (ueb *UIEventBuilder) Content(content any) *UIEventBuilder {
	ueb.newContent = content

	return ueb
}

func (ueb *UIEventBuilder) Position(pos UIRenderPosition) *UIEventBuilder {
	ueb.position = pos

	return ueb
}

func (ueb *UIEventBuilder) BuildWithError() (string, error) {
	if ueb.event == "" {
		return "", fmt.Errorf("Event cannot be null")
	}

	if ueb.targetNodeID == "" && ueb.event != UIEvent_ReplaceNodeContent {
		return "", fmt.Errorf("TargetNodeID cannot be null")
	}

	if ueb.receiverNodeSelector == "" {
		return "", fmt.Errorf("ReceiverNodeSelector cannot be null")
	}

	j, err := json.Marshal(struct {
		Event                              UIEvent
		TargetNodeID, ReceiverNodeSelector string
		NewContent                         any
		Position                           UIRenderPosition
	}{
		Event:                ueb.event,
		TargetNodeID:         ueb.targetNodeID,
		ReceiverNodeSelector: ueb.receiverNodeSelector,
		NewContent:           ueb.newContent,
		Position:             ueb.position,
	})
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func (ueb *UIEventBuilder) Build() string {
	if ueb.event == "" {
		log.Fatalln("Event cannot be null")
	}

	if ueb.targetNodeID == "" && ueb.event != UIEvent_ReplaceNodeContent {
		log.Fatalln("TargetNodeID cannot be null")
	}

	if ueb.receiverNodeSelector == "" {
		log.Fatalln("ReceiverNodeSelector cannot be null")
	}

	j, err := json.Marshal(struct {
		Event                              UIEvent
		TargetNodeID, ReceiverNodeSelector string
		NewContent                         any
		Position                           UIRenderPosition
	}{
		Event:                ueb.event,
		TargetNodeID:         ueb.targetNodeID,
		ReceiverNodeSelector: ueb.receiverNodeSelector,
		NewContent:           ueb.newContent,
		Position:             ueb.position,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return string(j)
}
