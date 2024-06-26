package SSEEvents

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/relepega/doujinstyle-downloader/internal/appUtils"
)

var (
	currentId = 0
	mu        sync.Mutex
)

type SSEMessage struct {
	Event string
	Data  string
	Id    int
}

func NewSSEMessage(data string) *SSEMessage {
	mu.Lock()
	defer mu.Unlock()

	currentId = currentId + 1

	return &SSEMessage{
		Event: "message",
		Data:  data,
		Id:    currentId,
	}
}

func NewSSEMessageWithEvent(event, data string) *SSEMessage {
	mu.Lock()
	defer mu.Unlock()

	currentId = currentId + 1

	return &SSEMessage{
		Event: event,
		Data:  data,
		Id:    currentId,
	}
}

func NewSSEMessageWithError(err error) *SSEMessage {
	mu.Lock()
	defer mu.Unlock()

	currentId = currentId + 1

	return &SSEMessage{
		Event: "error",
		Data:  err.Error(),
		Id:    currentId,
	}
}

func (m *SSEMessage) String() string {
	cleanData := appUtils.CleanString(m.Data)

	if m.Event == "" {
		return fmt.Sprintf("data: %s\nid: %d\n\n", cleanData, m.Id)
	}

	return fmt.Sprintf("event: %s\ndata: %s\nid: %d\n\n", m.Event, cleanData, m.Id)
}

type UIEvent string

const (
	NewNode            UIEvent = "new-node"
	ReplaceNode        UIEvent = "replace-node"
	ReplaceNodeContent UIEvent = "replace-node-content"
	RemoveNode         UIEvent = "remove-node"
)

type UIRenderPosition string

const (
	BeforeBegin UIRenderPosition = "beforebegin"
	AfterBegin  UIRenderPosition = "afterbegin"
	BeforeEnd   UIRenderPosition = "beforeend"
	AfterEnd    UIRenderPosition = "afterend"
	NullPos     UIRenderPosition = ""
)

type UIRenderEvent struct {
	Event        UIEvent          `json:"event"`        // "new-node" | "replace-node" | "replace-node-content" | "remove-node"
	TargetNode   string           `json:"targetNodeID"` // ID of the node that should be replaced
	ReceiverNode string           `json:"receiverNode"` // QuerySelector of the node that should receive the new content
	NewContent   string           `json:"newContent"`   // Newly rendered content: it can be either an entire DOM node or a DOM node content
	Position     UIRenderPosition `json:"position"`     // Position to where append content. Read: https://developer.mozilla.org/en-US/docs/Web/API/Element/insertAdjacentHTML
}

func NewUIRenderEvent(
	event UIEvent,
	targetNodeID, receiverNodeSelector, newContent string,
	position UIRenderPosition,
) *UIRenderEvent {
	return &UIRenderEvent{
		Event:        event,
		TargetNode:   targetNodeID,
		ReceiverNode: receiverNodeSelector,
		NewContent:   appUtils.CleanString(newContent),
		Position:     position,
	}
}

func (m *UIRenderEvent) String() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	// str := "{\"event\":\"replace-node\",\"targetNode\":\"65535\",\"receiverNode\":\"#content\",\"newNode\":\"\"}"

	return string(data), nil
}
