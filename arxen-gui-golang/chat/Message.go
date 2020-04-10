package chat

import (
	"github.com/rsocket/rsocket-go/payload"
	"time"
)

// universal interface for messages
type Message interface {
	MessageToPayload() payload.Payload
}

// text type message
type TextMessage struct {
	data string
	author string
	timestamp time.Time
}

// TODO implement me
func (t *TextMessage) MessageToPayload() payload.Payload {
	return payload.NewString("","")
}


func (t *TextMessage) Data() string {
	return t.data
}

func (t *TextMessage) SetData(data string) {
	t.data = data
}
