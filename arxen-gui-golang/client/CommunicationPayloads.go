package client

import "github.com/rsocket/rsocket-go/payload"

//noinspection ALL
const (
	CHAT_PARTICIPANTS_RESPONSE = "CHAT_PARTICIPANTS_RESPONSE"
	CHAT_PARTICIPANTS_REQUEST  = "CHAT_PARTICIPANTS_REQUEST"
)

type CommunicationPayload interface {
	ChatParticipantsResponse() payload.Payload
}

type CommunicationPayloadGenerator struct{}

func ChatParticipantsResponse(clientAddr string) payload.Payload {
	return payload.New([]byte(``), []byte(`{"source":"+", "type":"CHAT_PARTICIPANTS_RESPONSE"}`))
}
