package client

import (
	"context"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"log"
	"main/chat"
	"net"
	"os"
	"strings"
)

//  APP STACK:
//	+-------------------------+
//	|   Graphical Interface   |
//	+-------------------------+
//	|   Client daemon         |
//	+-------------------------+
//	|   Router daemon         |
//	+-------------------------+
//
//
// Client type
// basic tasks are:
// - communicate with routing daemon
// - directly cooperate with GUI
// - connect with other clients daemons
// - control every chat user is participating in

type Client struct {
	userIP          string
	clientsIPs      map[string]bool					// clientIP : status
	clientsSockets  map[rsocket.Client]string 		// socket : clientIP
	chatList        map[string]*chat.Chat 			// chatID, *Chat
	sendMessageList map[string]chan payload.Payload // payload and target chat format: map[clientIP] payload(message, chatID)

	secretKey 		string							// used for authentication
}

// return new User
func NewUser() *Client {
	// default port
	userAddr := "tcp://127.0.0.2:7878"

	// if everything is ok set userAddr as local IP
	if _userAddr, ok := GetOutboundIP(); ok {
		userAddr = "tcp://" + _userAddr.String() + ":7878"
		log.Println("NewUser: IP address = " + userAddr)
	} else {
		log.Println("NewUser: cannot obtain local IP address!")
	}

	// if there is env variable -> set clients ip to it
	if value, ok := os.LookupEnv("USER_ADDR"); ok {
		userAddr = value
		log.Println("NewUser: obtained predefined addr = " + userAddr)
	}


	// init channels
	_clientsIPs     := make(map[string]bool)
	_clientsSockets := make(map[rsocket.Client]string)
	_chatList       := make(map[string]*chat.Chat)
	_sendMessageList:= make(map[string]chan payload.Payload)

	return &Client{
		userIP:          userAddr,
		clientsIPs:      _clientsIPs,
		clientsSockets:  _clientsSockets,
		chatList:        _chatList,
		sendMessageList: _sendMessageList,
	}
}

// method listening and handling new connections to client
func (c *Client) eventListener() {
	// await for new connections
	err := rsocket.Receive().
		Resume().
		Fragment(1024).
		Acceptor(func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
			log.Println("eventListener: GOT REQUEST ", setup.DataUTF8())
			sendingSocket.OnClose(func(err error) {
				log.Println("eventListener: socket disconnected because ", err, " with ", setup.DataUTF8())
			})
			// returns custom handler
			return c.responder(setup), nil
		}).
		Transport(c.userIP).
		Serve(context.Background())
	panic(err)
}

// TODO implement till the end
// method used to create new chat
func (c *Client) createChat(initList []string) {

	chatIDstr := "123"

	// init new chat with complete users list
	// add userIP ex"tcp://10.5.0.2:7878" to that list
	tmpChat := chat.NewChat(chatIDstr, append(initList, c.userIP))

	// TODO fix me
	// go tmpChat.messagePrinter()

	// TODO TMP IMPLEMENTATION WARNING
	// not working if already connected to this user
	// get all users IP I want to connect
	for _, cli := range initList {
		c.clientsIPs[cli] = false
	}

	c.chatList[chatIDstr] = tmpChat

}

// used to obtain machine IP address
func GetOutboundIP() (net.IP, bool) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, true
}

func (c *Client) responder(setup payload.SetupPayload) rsocket.RSocket {
	// custom responder
	return rsocket.NewAbstractSocket(
		rsocket.MetadataPush(func(item payload.Payload) {
			log.Println("GOT METADATA_PUSH:", item)
		}),
		rsocket.FireAndForget(func(elem payload.Payload) {
			log.Println("GOT FNF:", elem)
		}),
		rsocket.RequestResponse(func(pl payload.Payload) mono.Mono {
			if meta, _ := pl.MetadataUTF8(); strings.EqualFold(meta, "REJECT_ME") {
				return nil
			}

			return mono.Just(pl)
		}),
		rsocket.RequestStream(func(pl payload.Payload) flux.Flux {
			s := pl.DataUTF8()
			m, _ := pl.MetadataUTF8()
			log.Println("data:", s, "metadata:", m)

			// handle getHosts request
			if dat, _ := pl.MetadataUTF8(); strings.EqualFold(dat, "CHAT_PARTICIPANTS_REQ") { // [chatID, REQ type]
				return flux.Create(func(ctx context.Context, emitter flux.Sink) {
					for _, ip := range c.chatList[pl.DataUTF8()].ClientsIPsList() {
						emitter.Next(payload.NewString(ip, "CHAT_PARTICIPANTS_RESP"))
					}
					emitter.Complete()
				})
			}

			return flux.Create(func(ctx context.Context, emitter flux.Sink) { emitter.Next(payload.NewString("EMPTY", "EMPTY")) })
		}),
		// TODO
		rsocket.RequestChannel(func(inputs rx.Publisher) flux.Flux {
			// control connected hosts:
			// get connecting hostIP and update user array

			// format: setup[clientIP]

			c.clientsIPs[setup.DataUTF8()] = true

			// create new chat
			c.createChat([]string{})

			inputs.(flux.Flux).DoFinally(func(s rx.SignalType) {
				log.Printf("signal type: %v", s)
				//close(receives)
			}).Subscribe(context.Background(), rx.OnNext(func(input payload.Payload) {
				log.Println("GOT MESSAGE: ", input.DataUTF8())
				tmpChatID, _ := input.MetadataUTF8()
				c.chatList[tmpChatID].MessagesChan <- input
			}))

			return flux.Create(func(ctx context.Context, s flux.Sink) {
				for mess := range c.sendMessageList[setup.DataUTF8()] {
					s.Next(mess)
				}
				s.Complete()
			})
		}),
	)
}