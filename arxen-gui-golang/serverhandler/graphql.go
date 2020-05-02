//go:generate go run github.com/99designs/gqlgen
package serverhandler

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"log"
	"main/client"
	"main/gql"
	"net/http"
	"sync"
	"time"
)

const GRAPHQL_ROUTE = "/graphql"

type ClientServer struct {
	client *client.Client
	mutex  sync.Mutex
}

func NewClientServer(client *client.Client) (*ClientServer, error) {
	return &ClientServer{
		client: client,
		mutex:  sync.Mutex{},
	}, nil
}

// serves graphql and vuejs (in future)
func (c *ClientServer) Serve(port int) error {
	mux := http.NewServeMux()
	mux.Handle(
		GRAPHQL_ROUTE,
		handler.GraphQL(gql.NewExecutableSchema(gql.Config{Resolvers: c}),
			handler.WebsocketUpgrader(websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}),
		),
	)
	mux.Handle("/playground", handler.Playground("GraphQL", GRAPHQL_ROUTE))

	// TODO add more routes

	handler := cors.AllowAll().Handler(mux)
	log.Println("Serving")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}

func (c *ClientServer) PostMessage(ctx context.Context, chatID string, text string) (*gql.TextMessage, error) {

	m := gql.TextMessage{
		ChatID:    chatID,
		User:      c.client.GetUserID(),
		TimeStamp: time.Now().UTC(),
		Text:      text,
	}

	c.mutex.Lock()
	c.client.GetChatList()[chatID].SendMessageChan <- m
	c.mutex.Unlock()

	// TODO add error handle
	return &m, nil
}

func (c *ClientServer) CreateChat(ctx context.Context, users []string) (*gql.Chat, error) {
	ch := c.client.CreateChat(users)

	tmpChat := &gql.Chat{
		ChatID:         ch.ChatID,
		ClientsIPsList: ch.ClientsIPsList(),
	}

	return tmpChat, nil
}

func (c *ClientServer) Messages(ctx context.Context, chatID string) ([]*gql.TextMessage, error) {
	// find chat and forward message
	c.mutex.Lock()
	textList := c.client.GetChatList()[chatID].TextMessageList
	c.mutex.Unlock()

	if textList != nil {
		return textList, nil
	}

	return nil, errors.New("text messages list could be not returned")
}

func (c *ClientServer) ChatUsers(ctx context.Context, chatID string) ([]string, error) {
	c.mutex.Lock()
	list := c.client.GetChatList()[chatID].ClientsIPsList()
	c.mutex.Unlock()

	if list != nil {
		return list, nil
	}

	return nil, errors.New("clients IPs List could be not returned")
}

func (c *ClientServer) Chats(ctx context.Context) ([]*gql.Chat, error) {
	panic("implement me")
}

func (c *ClientServer) MessagePosted(ctx context.Context, chatID string) (<-chan *gql.TextMessage, error) {
	// Create new channel for request
	c.mutex.Lock()
	messages := c.client.GetChatList()[chatID].MessagesChan
	c.mutex.Unlock()

	return messages, nil
}

func (c *ClientServer) UserJoined(ctx context.Context, chatID string) (<-chan string, error) {
	// TODO implement me
	return make(chan string, 1), nil
}

func (c *ClientServer) ChatCreated(ctx context.Context) (<-chan *gql.Chat, error) {
	// TODO implement me
	return make(chan *gql.Chat, 1), nil
}

func (c *ClientServer) Mutation() gql.MutationResolver {
	return c
}

func (c *ClientServer) Query() gql.QueryResolver {
	return c
}

func (c *ClientServer) Subscription() gql.SubscriptionResolver {
	return c
}
