package chat

import (
	"github.com/rsocket/rsocket-go/payload"
	"time"
)

// universal interface for messages
type Message interface {
	MessageToPayload() payload.Payload
	MessageToJsonString() string
}

// text type message
type TextMessage struct {
	Data      string
	Author    string
	Timestamp time.Time
}

// TODO implement me
func (t *TextMessage) MessageToPayload() payload.Payload {
	return payload.New([]byte(t.Data), []byte(`{"Author":"`+t.Author+`","Timestamp":"`+t.Timestamp.String()+`"}`))
}

// TODO implement me
func (t *TextMessage) MessageToJsonString() string {
	return ""
}

// returns new TextMessage based on payload
func PayloadToTextMessage(payl payload.Payload) *TextMessage {
	// TODO implement me
	return &TextMessage{
		Data:      payl.DataUTF8(),
		Author:    "tmp",
		Timestamp: time.Now(),
	}
}