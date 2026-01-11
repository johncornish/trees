package protocol

import (
	"encoding/json"
	"trees/internal/domain"
)

// Message types for the TCP protocol.
const (
	TypeSubscribe   = "subscribe"
	TypeSubscribed  = "subscribed"
	TypePublishTree = "publishTree"
	TypeTreeAdded   = "treeAdded"
)

// BaseMessage contains the common type field.
type BaseMessage struct {
	Type string `json:"type"`
}

// SubscribeMessage is sent by a client to subscribe to a project.
type SubscribeMessage struct {
	Type       string `json:"type"`
	ProjectKey string `json:"projectKey"`
}

// SubscribedMessage is sent by the server to confirm subscription.
type SubscribedMessage struct {
	Type       string `json:"type"`
	ProjectKey string `json:"projectKey"`
}

// PublishTreeMessage is sent by a client to publish a new tree.
type PublishTreeMessage struct {
	Type       string           `json:"type"`
	ProjectKey string           `json:"projectKey"`
	Tree       domain.TaskTree  `json:"tree"`
}

// TreeAddedMessage is broadcast by the server to all subscribers.
type TreeAddedMessage struct {
	Type       string           `json:"type"`
	ProjectKey string           `json:"projectKey"`
	Tree       domain.TaskTree  `json:"tree"`
}

// ParseMessageType extracts the message type from JSON data.
func ParseMessageType(data []byte) (string, error) {
	var base BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return "", err
	}
	return base.Type, nil
}
