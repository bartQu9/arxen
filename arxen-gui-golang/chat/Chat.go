package chat

import (
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
)

type Chat struct {
	chatID         string
	clientsIPsList []string
	MessagesChan   chan payload.Payload
	listiner       interface{}
	f              flux.Flux
}

// Create new chat
// args - chatID: ID of chat (numeric string); clientsIPsList: list of other participants addresses (list of strings)
//
func NewChat(chatID string, clientsIPsList []string) *Chat {
	return &Chat{chatID: chatID, clientsIPsList: clientsIPsList, MessagesChan: make(chan payload.Payload)}
}

func (c Chat) ClientsIPsList() []string {
	return c.clientsIPsList
}