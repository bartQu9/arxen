//go:generate go run github.com/99designs/gqlgen
package serverhandler

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"main/chat"
	"main/client"
	"main/gql"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
)

// Route directory for grapql server api
const GRAPHQL_ROUTE = "/graphql"

// struct combining client with mutex
type ClientServer struct {
	client *client.Client
	mutex  sync.Mutex
}

// NewChatLastMessage implement me
func (c *ClientServer) NewChatLastMessage(ctx context.Context, chatID string) (<-chan *string, error) {
	panic("implement me")
}

// NewFriend implement me
func (c *ClientServer) NewFriend(ctx context.Context) (<-chan *gql.Friend, error) {
	panic("implement me")
}

// GetFriendsTypeList returns friends of user as string Friend struct
func (c *ClientServer) GetFriendsTypeList(ctx context.Context) ([]*gql.Friend, error) {
	var tmpFriendsList []*gql.Friend

	// find chat and forward message
	c.mutex.Lock()
	for _, friend := range c.client.FriendsList {
		tmpFriendsList = append(tmpFriendsList, friend)
	}
	c.mutex.Unlock()

	//log.Println("FetchMessages: chatID ", chatID, " resp: ", textList)

	log.WithFields(log.Fields{
		"textList": tmpFriendsList,
	}).Debug("GetFriendsTypeList:")

	if tmpFriendsList != nil {
		return tmpFriendsList, nil
	}

	//return nil, errors.New("text messages list could be not returned")
	return []*gql.Friend{}, nil
}

// ChangeChatName implement me
func (c *ClientServer) ChangeChatName(ctx context.Context, chatID string, chatName string) (*string, error) {
	panic("implement me")
}

// GetFriendList returns friends of user as string list
func (c *ClientServer) GetFriendList(ctx context.Context) ([]*string, error) {
	var friendsStringList []*string

	// map each friend to name (in future {name, userID})
	c.mutex.Lock()
	for _, friend := range c.client.FriendsList {
		log.Debug("GetFriendList: having ", friend)
		tmpStr := friend.Nick
		friendsStringList = append(friendsStringList, tmpStr)
	}
	c.mutex.Unlock()

	log.WithFields(log.Fields{
		"friendsStringList": friendsStringList,
	}).Debug("GetFriendList: responded to request")

	return friendsStringList, nil
}

// GetUserName returns user name
func (c *ClientServer) GetUserName(ctx context.Context) (string, error) {
	return c.client.GetUserID(), nil
}

// AddFriend implement me
func (c *ClientServer) AddFriend(ctx context.Context, userUUID string) (*string, error) {
	panic("implement me")
}

// ChangeNick implement me
func (c *ClientServer) ChangeNick(ctx context.Context, userNick string) (*string, error) {
	panic("implement me")
}

// ClientWritingAlert implement me
func (c *ClientServer) ClientWritingAlert(ctx context.Context, chatID string) (<-chan *string, error) {
	panic("implement me")
}

// FetchMessages returns numOfMessages messages from particular chat
func (c *ClientServer) FetchMessages(ctx context.Context, chatID string, numOfMessages int) ([]*gql.TextMessage, error) {
	// find chat and forward message
	c.mutex.Lock()
	numOfExistingMessages := len(c.client.GetChatList()[chatID].TextMessageList)
	// make sure not exiting number of map array elements
	if numOfMessages > numOfExistingMessages {
		numOfMessages = numOfExistingMessages
	}
	textList := c.client.GetChatList()[chatID].TextMessageList[0:numOfMessages]
	c.mutex.Unlock()

	//log.Println("FetchMessages: chatID ", chatID, " resp: ", textList)

	log.WithFields(log.Fields{
		"chatID": chatID,
		"textList": textList,
	}).Debug("FetchMessages:")

	if textList != nil {
		return textList, nil
	}

	//return nil, errors.New("text messages list could be not returned")
	return []*gql.TextMessage{}, nil
}

// ClientWriting implement me
func (c *ClientServer) ClientWriting(ctx context.Context, chatID string, userID string) (*string, error) {
	panic("implement me")
}

// ChangeChatAvatar implement me
func (c *ClientServer) ChangeChatAvatar(ctx context.Context, chatID string, avatarAddr string) (*string, error) {
	panic("implement me")
}

// NewClientServer returns new ClientServer
func NewClientServer(client *client.Client) (*ClientServer, error) {
	return &ClientServer{
		client: client,
		mutex:  sync.Mutex{},
	}, nil
}

// Serve serves graphql and vuejs (in future)
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

	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(`Hello`))
		if err != nil {
			log.WithError(err).Error("Serve:")
		}
	})

	// TODO refactor to configuration
	fileServer := http.FileServer(http.Dir("public/resources"))
	mux.Handle("/static/", http.StripPrefix(strings.TrimRight("/static/", "/"), fileServer))

	// TODO add more routes

	handler := cors.AllowAll().Handler(mux)
	log.Info("Serving")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}

// PostMessage is mutation used to post new message on chat
func (c *ClientServer) PostMessage(ctx context.Context, chatID string, text string) (*gql.TextMessage, error) {

	m := gql.TextMessage{
		MessageID: ksuid.New().String(),
		ChatID:    chatID,
		User:      c.client.GetUserID(),
		TimeStamp: time.Now().UTC(),
		Text:      text,
	}

	c.mutex.Lock()
	c.client.GetChatList()[chatID].SendMessageChan <- m
	c.mutex.Unlock()

	//log.Println("PostMessage: chatID ", chatID, " text \"", text, "\", resp: ", m)

	log.WithFields(log.Fields{
		"chatID": chatID,
		"resp": text,
	}).Debug("PostMessage:")

	// TODO add error handle
	return &m, nil
}

// CreateChat is mutation creating new chat based on users list
func (c *ClientServer) CreateChat(ctx context.Context, users []string) (*gql.Chat, error) {
	ch := c.client.CreateChat(users)

	tmpChat := &gql.Chat{
		ChatID:         ch.ChatID,
		ClientsIPsList: ch.ClientsIPsList(),
	}

	//log.Println("CreateChat: users ", users, " resp: ", tmpChat.ChatID)

	log.WithFields(log.Fields{
		"users": users,
		"resp": tmpChat.ChatID,
	}).Debug("CreateChat:")


	return tmpChat, nil
}

// Messages is query returns all messages from particular chat
func (c *ClientServer) Messages(ctx context.Context, chatID string) ([]*gql.TextMessage, error) {
	// find chat and forward message
	c.mutex.Lock()
	textList := c.client.GetChatList()[chatID].TextMessageList
	c.mutex.Unlock()

	// log.Println("Messages: chatID ", chatID, " resp: ", textList)

	log.WithFields(log.Fields{
		"chatID": chatID,
		"resp": textList,
	}).Debug("Messages:")

	if textList != nil {
		return textList, nil
	}

	//return nil, errors.New("text messages list could be not returned")
	return []*gql.TextMessage{}, nil
}

// ChatUsers is query that returns chat users
func (c *ClientServer) ChatUsers(ctx context.Context, chatID string) ([]string, error) {
	c.mutex.Lock()
	list := c.client.GetChatList()[chatID].ClientsIPsList()
	c.mutex.Unlock()

	// log.Println("ChatUsers: chatID ", chatID, " resp: ", list)

	log.WithFields(log.Fields{
		"chatID": chatID,
		"resp": list,
	}).Debug("ChatUsers:")

	if list != nil {
		return list, nil
	}

	return nil, errors.New("clients IPs List could be not returned")
}

// Chats is query that return chat list
func (c *ClientServer) Chats(ctx context.Context) ([]*gql.Chat, error) {
	var chats map[string]*chat.Chat
	var gqlChats []*gql.Chat

	c.mutex.Lock()
	chats = c.client.GetChatList()
	c.mutex.Unlock()

	for _, ch := range chats {
		var lastMessage *gql.TextMessage
		if len(ch.TextMessageList) != 0 {
			lastMessage = ch.TextMessageList[len(ch.TextMessageList)-1]
		}
		gqlChats = append(gqlChats, &gql.Chat{
			ChatID:         ch.ChatID,
			ClientsIPsList: ch.ClientsIPsList(),
			LatestMessage: 	lastMessage,
			ClientWriting:  nil,
			ChatName:       &ch.ChatName,
		})
	}

	// log.Println("Chats: resp: ", gqlChats)

	log.WithFields(log.Fields{
		"resp": gqlChats,
	}).Debug("Chats:")


	return gqlChats, nil
}

// MessagePosted is subscription event when new message is posted in particular chat
func (c *ClientServer) MessagePosted(ctx context.Context, chatID string) (<-chan *gql.TextMessage, error) {
	// Create new channel for request
	c.mutex.Lock()
	messages := c.client.GetChatList()[chatID].MessagesChan
	c.mutex.Unlock()

	// log.Println("MessagePosted: chatID ", chatID, " resp: ", messages)

	log.WithFields(log.Fields{
		"chatID": chatID,
		"resp": messages,
	}).Debug("MessagePosted:")

	return messages, nil
}

// UserJoined is subscription event when new user joins chat
func (c *ClientServer) UserJoined(ctx context.Context, chatID string) (<-chan string, error) {
	// TODO implement me
	// log.Println("UserJoined: chatID ", chatID)

	log.WithFields(log.Fields{
		"chatID": chatID,
	}).Debug("UserJoined:")

	return make(chan string, 1), nil
}

// ChatCreated is subscription event when new chat is created
// to be implemented
func (c *ClientServer) ChatCreated(ctx context.Context) (<-chan *gql.Chat, error) {
	// TODO implement me
	log.Println("ChatCreated: ")
	return make(chan *gql.Chat, 1), nil
}

// Mutation returns mutation resolver
func (c *ClientServer) Mutation() gql.MutationResolver {
	return c
}

// Query returns query resolver
func (c *ClientServer) Query() gql.QueryResolver {
	return c
}

// Subscription returns subscription resolver
func (c *ClientServer) Subscription() gql.SubscriptionResolver {
	return c
}
