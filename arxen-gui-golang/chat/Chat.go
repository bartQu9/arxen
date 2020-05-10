package chat

import (
	"github.com/rsocket/rsocket-go/rx/flux"
	"main/gql"
)

type Chat struct {

	// UUID for chat
	ChatID string

	// list of all participating in chat Clients
	clientsIPsList []string

	// all messages within the chat goes here
	MessagesChan chan *gql.TextMessage

	// messages sent by Client goes here
	SendMessageChan chan gql.TextMessage

	// list of all messages (in database in the future)
	TextMessageList []*gql.TextMessage

	listiner interface{}
	f        flux.Flux

	// socket connected with to frontend
	// socket *websocket.Conn
}

// TODO add logic for adding clients from friends list

// Create new chat
// args - ChatID: ID of chat (numeric string); clientsIPsList: list of other participants addresses (list of strings)
//
func NewChat(chatID string, clientsIPsList []string) *Chat {
	return &Chat{ChatID: chatID, TextMessageList: []*gql.TextMessage{}, clientsIPsList: clientsIPsList, MessagesChan: make(chan *gql.TextMessage, 100), SendMessageChan: make(chan gql.TextMessage)}
}

func (c Chat) ClientsIPsList() []string {
	return c.clientsIPsList
}

func (c *Chat) stopChat() {}

//const (
//	socketBufferSize  = 1024
//	messageBufferSize = 256
//)

//var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

//func (c *Chat) read() {
//	defer c.socket.Close()
//	for {
//		//var msg *string
//		//
//		//err := c.socket.ReadJSON(&msg)
//		//log.Println("read(): Got Message from Websocket ", *msg)
//		//if err != nil {
//		//	return
//		//}
//		//
//		//msgToSend := TextMessage{
//		//	Data: *msg,
//		//	Timestamp: time.Now(),
//		//	Author: "tmp_solution",
//		//}
//		//// msgToSend.Data = *msg
//		//// msgToSend.Timestamp = time.Now()
//		//// TODO solve stamping messages with author
//		//// msgToSend.Author = "tmp_solution"
//		//// to payload
//		//c.SendMessageChan <- msgToSend
//		var msg *gql.TextMessage
//		err := c.socket.ReadJSON(&msg)
//		if err != nil {
//			return
//		}
//		msg.TimeStamp = time.Now()
//		msg.User = "tmp"
//		log.Println("read(): Got Message from Websocket ", *msg)
//		c.SendMessageChan <- *msg
//	}
//}

//func (c *Chat) write() {
//	defer c.socket.Close()
//	for msg := range c.MessagesChan {
//		// TODO fix me properly
//		// implement proper type transition
//		log.Println("write(): message to be written: ", msg)
//		//err := c.socket.WriteJSON("{\"Data\":\""+msg.DataUTF8()+"\",\"Author\":\""+"TBD"+"\"}")
//		err := c.socket.WriteJSON(msg)
//		if err != nil {
//			return
//		}
//	}
//}

//// probably removed in future commits
//func (c *Chat) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//	socket, err := upgrader.Upgrade(w, req, nil)
//	c.socket = socket // no idea if it works actually
//	if err != nil {
//		log.Fatal("ServeHTTP:", err)
//		return
//	}
//
//	defer func() { c.stopChat() }()
//	go c.write()
//	c.read()
//}
