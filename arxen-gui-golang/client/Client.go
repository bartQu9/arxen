package client

import (
	"context"
	"encoding/json"
	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
	"log"
	"main/chat"
	"main/gql"
	"net"
	"os"
	"strings"
	"sync"
	"time"
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

// rate of refreshing connections with other clients
const CONNECTIONS_UPDATE_REFRESH_RATE = 10 * time.Second

// Client: basic struct handling connections between other clients
type Client struct {
	userIP     string
	clientsIPs map[string]bool // clientIP : status
	// not in use rn
	clientsSockets      map[rsocket.Client]string       // socket : clientIP
	chatList            map[string]*chat.Chat           // chatID, *Chat
	sendDataList        map[string]chan payload.Payload // payload and target chat format: map[clientIP] payload(message, chatID)
	receivedPayloadChan chan payload.Payload            // channel with all incoming payloads

	FriendsList map[string]*gql.Friend // map[friendsNick]Friend
	secretKey   string            // used for authentication

	mutex 		sync.Mutex			// to prevent access to same data by two goroutines
}

// GetChatList returns chat list map
func (c *Client) GetChatList() map[string]*chat.Chat {
	return c.chatList
}

// NewClient returns new Client
func NewClient() *Client {
	// default port
	userAddr := "tcp://127.0.0.2:7878"

	// if everything is ok set userAddr as local IP
	if _userAddr, ok := GetOutboundIP(); ok {
		userAddr = "tcp://" + _userAddr.String() + ":7878"
		log.Println("NewClient: IP address = " + userAddr)
	} else {
		log.Println("NewClient: cannot obtain local IP address!")
	}

	// if there is env variable -> set clients ip to it
	if value, ok := os.LookupEnv("USER_ADDR"); ok {
		userAddr = value
		log.Println("NewClient: obtained predefined addr = " + userAddr)
	}

	// init channels
	_clientsIPs := make(map[string]bool)
	_clientsSockets := make(map[rsocket.Client]string)
	_chatList := make(map[string]*chat.Chat)
	_sendMessageList := make(map[string]chan payload.Payload)
	_receivedPayloadChan := make(chan payload.Payload)
	_FriendsList := make(map[string]*gql.Friend)

	return &Client{
		userIP:              userAddr,
		clientsIPs:          _clientsIPs,
		clientsSockets:      _clientsSockets,
		chatList:            _chatList,
		sendDataList:        _sendMessageList,
		receivedPayloadChan: _receivedPayloadChan,
		FriendsList:		 _FriendsList,
	}
}

// eventListener is method listening and handling new connections to client
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

// CreateChat method is used to create new chat
// TODO implement till the end
func (c *Client) CreateChat(initList []string) *chat.Chat {

	chatIDstr := uuid.New().String()

	logger.WithField("chatID", chatIDstr).Info("CreateChat: creating new chat")

	// init new chat with complete users list
	// add userIP ex"tcp://10.5.0.2:7878" to that list
	tmpChat := chat.NewChat(chatIDstr, append(initList, c.userIP))

	// TODO fix me
	// go tmpChat.messagePrinter()

	// TODO TMP IMPLEMENTATION WARNING
	// not working if already connected to this user
	// get all users IP I want to connect
	for _, cli := range initList {
		c.mutex.Lock()
		if _, ok := c.clientsIPs[cli]; !ok {
			c.clientsIPs[cli] = false
		}
		c.mutex.Unlock()
	}

	go c.chatMessagesHandler(tmpChat)

	c.mutex.Lock()
	c.chatList[chatIDstr] = tmpChat
	c.mutex.Unlock()

	// advert new chat
	c.receivedPayloadChan <- payload.New([]byte(chatIDstr), c.getMetadataTag(CHAT_ADVERT_REQUEST))

	return tmpChat

	// CODE BELOW NOT NEEDED;
	// TODO REMOVE IN FUTURE
	// create map of adv statuses
	//tmpAdList := initList

	/*
		// do while chat is not adv to all clients

		go func() {
			for {
				for i, addr := range tmpAdList {
					if status := c.clientsIPs[addr]; status {
						// delete record
						tmpAdList = append(tmpAdList[:i], tmpAdList[i+1:]...)
					}
				}
				if len(tmpAdList) == 0 {
					break
				}
			}
		}()
	*/
}

// createSlaveChat is version of CreateChat used when chatID is already known
func (c *Client) createSlaveChat(initList []string, chatIDstr string) {
	// init new chat with complete users list
	// add userIP ex"tcp://10.5.0.2:7878" to that list
	tmpChat := chat.NewChat(chatIDstr, append(initList, c.userIP))

	// TODO fix me
	// go tmpChat.messagePrinter()

	// TODO TMP IMPLEMENTATION WARNING
	// not working if already connected to this user
	// get all users IP I want to connect
	for _, cli := range initList {
		c.mutex.Lock()
		if _, ok := c.clientsIPs[cli]; !ok {
			c.clientsIPs[cli] = false
		}
		c.mutex.Unlock()
	}

	go c.chatMessagesHandler(tmpChat)

	c.mutex.Lock()
	c.chatList[chatIDstr] = tmpChat
	c.mutex.Unlock()

	log.Println("createSlaveChat: Created new Chat")

}

// connectionsHandler is a handler of all connections across itself and other clients
func (c *Client) connectionsHandler() {
	for {
		// refresh at rate
		time.Sleep(CONNECTIONS_UPDATE_REFRESH_RATE)

		for addr, status := range c.clientsIPs {
			// if client not connected to particular client try to connect
			// find if chan for that client exists
			// TODO can be written better
			if c.sendDataList[addr] == nil {
				log.Println("connectionsHandler: chan non existing - creating ", addr)
				c.mutex.Lock()
				ch := make(chan payload.Payload)
				c.sendDataList[addr] = ch
				c.mutex.Unlock()
			}
			if !status {
				go c.connectToClient(c.sendDataList[addr], addr)
				// after finished update record
				c.mutex.Lock()
				c.clientsIPs[addr] = true
				c.mutex.Unlock()
			}
		}
	}
}

// connectToClient periodically check if client is connected to desired clients
// Possible type problem: struct vs payload
func (c *Client) connectToClient(ch chan payload.Payload, addr string) {
	// goroutine for connecting to clients
	// handle channels

	// in advanced scenario ask host for chat clients ips

	// new client
	// TODO change literals to constants
	cli, err := rsocket.
		Connect().
		SetupPayload(payload.NewString(c.userIP, "1234")).
		Resume().
		Fragment(1024).
		OnClose(func(err error) {
			log.Println("connectToClient: connection with ", addr, " closed because ", err)
			c.clientsIPs[addr] = false
		}).
		Transport(addr).
		Start(context.Background())
	if err != nil {
		logger.WithField("err", err).Warn("Connection Error occurred")
		switch err.Error() {
		case "Connection Error occurred":
			logger.WithField("err", err.Error()).Warn("connectToClient: connection was not established")
			return
		default:
			panic(err)
		}
	}

	defer cli.Close()

	// create tmp flux
	// TODO problem: who is the target
	// TODO add option of sending custom messages
	// TODO make this flux never cancel!
	f := flux.Create(func(ctx context.Context, s flux.Sink) {
		//log.Println("STARTED sending new message")
		for mess := range ch {
			//log.Println("SENDING new message")
			s.Next(mess)
		}
		log.Println("connectToClient: transmission completed")
		s.Complete()
	}).DoFinally(func(s rx.SignalType) {
		log.Println("connectToClient: GOT SIGNAL ", s)
	})


	log.Println("REQUESTING CHANNEL WITH ", addr)

	// possible error
	// TODO remove debug stats
	_, err = cli.RequestChannel(f).
		DoOnNext(func(elem payload.Payload) {
			log.Println("GOT new message")
			//tmpChatID, _ := elem.MetadataUTF8()
			// TODO check if fixed
			//c.chatList[tmpChatID].MessagesChan <- elem
			c.receivedPayloadChan <- elem
		}).DoOnComplete(func() {
			log.Println("connectToClient: job completed")
		}).DoOnError(func(e error) {
			log.Println("connectToClient: ERROR occurred ", e)
		}).DoFinally(func(s rx.SignalType) {
			log.Println("connectToClient: finally ", s)
		}).
		BlockLast(context.Background())
}

// clientManager is not in use at this moment
// runs eventListener() and manages connections
func (c *Client) clientManager() {

}

// GetUserID returns userIP/ID
// TODO solve userID/IP
func (c *Client) GetUserID() string {
	// TODO change in the future
	return c.userIP
}

// GetOutboundIP can be used to obtain machine IP address
func GetOutboundIP() (net.IP, bool) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, true
}

// responder is factory for rsocket.RSocket instance
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
			// s := pl.DataUTF8()
			// m, _ := pl.MetadataUTF8()
			// log.Println("data:", s, "metadata:", m)

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

			if c.sendDataList[setup.DataUTF8()] == nil {
				log.Println("responder: chan non existing - creating ", setup.DataUTF8())
				ch := make(chan payload.Payload)
				c.sendDataList[setup.DataUTF8()] = ch
			}

			// TODO possibly remove

			// create new chat
			//c.CreateChat([]string{})

			inputs.(flux.Flux).DoFinally(func(s rx.SignalType) {
				log.Printf("responder: signal type: %v", s)
				//close(receives)
			}).SubscribeOn(scheduler.Elastic()).DoOnError(func(e error) {
				log.Println("responder: ERROR ", e)
			}).Subscribe(context.Background(), rx.OnNext(func(input payload.Payload) {

				// TODO sort messages

				// log.Println(input)

				// TODO FIX ME ERROR HERE: "runtime error: invalid memory address or nil pointer dereference"
				log.Println("responder: GOT MESSAGE: ", input.DataUTF8())
				// tmpChatID, _ := input.MetadataUTF8()
				// c.chatList[tmpChatID].MessagesChan <- input

				c.receivedPayloadChan <- input
			}))

			return flux.Create(func(ctx context.Context, s flux.Sink) {
				for mess := range c.sendDataList[setup.DataUTF8()] {
					s.Next(mess)
				}
				s.Complete()
			}).DoFinally(func(s rx.SignalType) {
				log.Println("responder: Got signal ", s)
			})
		}),
	)
}

// payloads:
// CHAT_MESSAGE:			  {message,{source, type, chatID}}
// CHAT_PARTICIPANTS_REQUEST: {chatID, {source, type}}

// receivedPayloadHandler is helper, handling all incoming messages from each connection
func (c *Client) receivedPayloadHandler() {
	// this "for" is basically onNext()
	for payl := range c.receivedPayloadChan {

		// read message data/metadata
		// based on input do something
		metaByteJson, _ := payl.Metadata()
		var metadata map[string]interface{}
		if err := json.Unmarshal(metaByteJson, &metadata); err != nil {
			// TODO better handle error
			panic(err)
		}

		// TODO add authentication process for request (client not participating in chat can get its participants)

		// TODO implement me till the end
		// log.Println("receivedPayloadHandler: INCOMING: ", payl)

		logger.WithField("payl", payl).Trace("receivedPayloadHandler: INCOMING")

		switch metadata["type"].(string) {
		case CHAT_MESSAGE:
			// TODO handle incoming messages
			// the source
			// authentication
			if dest := metadata["chatId"]; dest != nil {
				// send to appropriate chat
				tmpTextMessage := PayloadToGraphqlTextMessage(payl)
				c.chatList[dest.(string)].MessagesChan <- &tmpTextMessage
				logger.Trace("receivedPayloadHandler: After CHAN")
				c.chatList[dest.(string)].TextMessageList = append(c.chatList[dest.(string)].TextMessageList, &tmpTextMessage)
				//<- chat.TextMessage{
				//	Data:      payl.DataUTF8(),
				//	Author:    metadata["source"].(string),
				//	Timestamp: time.Now(),
				//}

				logger.Trace("receivedPayloadHandler: Left CHAT_MESSAGE section")
			}
		case CHAT_PARTICIPANTS_REQUEST:
			// send all participating clients IPs to requester
			// v1
			//for _, addr := range c.chatList[payl.DataUTF8()].ClientsIPsList() {
			//	tmpSendPayl := payload.New([]byte(addr), c.getMetadataTag(CHAT_PARTICIPANTS_RESPONSE))
			//	c.sendDataList[metadata["source"].(string)] <- tmpSendPayl
			//}
			//v2 possible problem is size limit of payload
			log.Println("receivedPayloadHandler: got new CHAT_PARTICIPANTS_REQUEST")
			if _, ok := c.chatList[payl.DataUTF8()]; !ok {
				logger.Warn("receivedPayloadHandler: chatID not found in clients chatList")
				break
			}
			addrString := strings.Join(c.chatList[payl.DataUTF8()].ClientsIPsList(), ",")
			c.sendDataList[metadata["source"].(string)] <- payload.New([]byte(addrString), c.getMetadataTag(CHAT_PARTICIPANTS_RESPONSE, payl.DataUTF8()))
			log.Println("receivedPayloadHandler: sending chat CHAT_PARTICIPANTS_RESPONSE")

		case CHAT_ADVERT_REQUEST:
			// phantom request
			// should work :/
			for _, addr := range c.chatList[payl.DataUTF8()].ClientsIPsList() {
				if addr != c.userIP {
					// check if corresponding chan exists
					if c.sendDataList[addr] == nil {
						log.Println("receivedPayloadHandler: chan non existing - creating ", addr)
						// tmp solution
						ch := make(chan payload.Payload)
						c.sendDataList[addr] = ch
					}
					// send to each chan CHAT_ADVERT
					c.sendDataList[addr] <- payload.New(payl.Data(), c.getMetadataTag(CHAT_ADVERT, payl.DataUTF8()))
				}
			}
		case CHAT_ADVERT:
			// ask for all participants
			c.sendDataList[metadata["source"].(string)] <- payload.New(payl.Data(), c.getMetadataTag(CHAT_PARTICIPANTS_REQUEST))
			log.Println("receivedPayloadHandler: asking by CHAT_PARTICIPANTS_REQUEST")

		case CHAT_PARTICIPANTS_RESPONSE:
			// create new chat
			// not ideal solution
			addrArray := strings.Split(payl.DataUTF8(), ",")
			log.Println("receivedPayloadHandler: beginning creation of new chat")
			c.createSlaveChat(addrArray, metadata["chatID"].(string))
		default:
			log.Println("ERROR! UNSUPPORTED PAYLOAD METADATA TYPE")
		}

	}
}

// chatMessagesHandler handles forwarding messages from particular chat
func (c *Client) chatMessagesHandler(chat *chat.Chat) {
	for newMessageToBeSend := range chat.SendMessageChan {

		// transform message
		payloadMessage := payload.New([]byte(newMessageToBeSend.Text), c.getMetadataTag(CHAT_MESSAGE, chat.ChatID, newMessageToBeSend.User, newMessageToBeSend.TimeStamp.String(), newMessageToBeSend.MessageID))

		// forward to oneself
		c.receivedPayloadChan <- payloadMessage

		log.Println("chatMessagesHandler: Message to be send: ", payloadMessage)

		// forward to all connected hosts
		for _, clientIP := range chat.ClientsIPsList() {
			if clientIP != c.userIP {
				c.sendDataList[clientIP] <- payloadMessage
			}
		}
	}
}

// getMetadataTag: function returning metadata for payload
// args:
// args[0]: type of request/response to be generated
// args[1:]: extra arguments:
func (c *Client) getMetadataTag(args ...string) []byte {
	switch args[0] {
	case CHAT_PARTICIPANTS_REQUEST:
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `"}`)
	case CHAT_PARTICIPANTS_RESPONSE:
		// args[1]: chatID
		if len(args) < 2 {
			panic("getMetadataTag: Too few arguments")
		}
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `","chatID":"` + args[1] + `"}`)
	case CHAT_MESSAGE:
		// args[1]: chatID, args[2]: user, args[3]: timeStamp, args[4]: MessageID
		if len(args) < 5 {
			panic("getMetadataTag: Too few arguments")
		}
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `","chatId":"` + args[1] + `", "user":"` + args[2] + `", "timeStamp":"` + args[3] + `", "MessageID": "`+ args[4] +`"}`)
	case CHAT_ADVERT_REQUEST:
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `"}`)
	case CHAT_ADVERT:
		// args[1]: chatName
		if len(args) < 2 {
			panic("getMetadataTag: Too few arguments")
		}
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `", "chatName": "` + args[1] + `"}`)
	default:
		log.Fatalln("getMetadataTag: Bad message type")
		return nil
	}
}

// PayloadToGraphqlTextMessage converts incoming payload to TextMessage (defined in gql module)
// CHAT_MESSAGE:			  {text,{source, type, chatID/chatId, user, timeStamp}}
// where text is part of TextMessage
func PayloadToGraphqlTextMessage(p payload.Payload) gql.TextMessage {
	// TODO better solution for escaping json

	tmpMetadata, _ := p.Metadata()
	var metadata map[string]interface{}
	//
	//dataJson, err := strconv.Unquote(string(tmpJson))
	//if err != nil {
	//	logger.WithError(err).Fatal("PayloadToGraphqlTextMessage: dataJson")
	//}

	if err := json.Unmarshal(tmpMetadata, &metadata); err != nil {
		panic(err)
	}

	// TODO what if empty data

	date, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", metadata["timeStamp"].(string))
	if err != nil {
		log.Fatalln(err)
	}

	var chatID string

	if chatID :=  metadata["chatId"].(string); len(chatID) == 0 {
		chatID = metadata["chatID"].(string)
	}

	return gql.TextMessage{
		MessageID: metadata["MessageID"].(string),
		ChatID:    chatID,
		User:      metadata["user"].(string),
		TimeStamp: date,
		Text:      p.DataUTF8(),
	}
}

// GraphqlTextMessageToByte converts text message format to bytes
// probably redundant in the future
func GraphqlTextMessageToByte(message gql.TextMessage) []byte {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Fatalln("GraphqlTextMessageToByte: ", err)
	}
	log.Println("GraphqlTextMessageToByte: ",jsonMessage)
	return []byte(`{"chatId": "` + message.ChatID + `", "user": "` + message.User + `", "timeStamp": "` + message.TimeStamp.String() + `", "text": "` + message.Text + `" }`)
}

// --------------------------------------------------------
// web part
// --------------------------------------------------------

//const localServerAddress = ":7879"

//type templateHandler struct {
//	once     sync.Once
//	filename string
//	templ    *template.Template
//	data     map[string]interface{}
//}

//func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	t.once.Do(func() {
//		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
//	})
//
//	//data := map[string]interface{}{
//	//	"Host": r.Host,
//	//}
//
//	t.templ.Execute(w, t.data)
//}
//
//func (c *Client) HttpServer() {
//	http.HandleFunc("/testing", func(writer http.ResponseWriter, request *http.Request) {
//
//	})
//	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
//		// TODO implement me
//		writer.Write([]byte(`Hello TODZIALA`))
//	})
//	http.Handle("/fixme", &templateHandler{filename: "chat.html"})
//	http.HandleFunc("/room/", func(writer http.ResponseWriter, request *http.Request) {
//		segs := strings.Split(request.URL.Path, "/")
//		urlChatID := segs[len(segs)-1]
//
//		if c.chatList[urlChatID] == nil {
//			writer.WriteHeader(http.StatusBadRequest)
//			// TODO better handle error
//			log.Fatalln("Bad URL: chat seems to not exist")
//			// panic("Bad URL")
//		}
//
//		c.chatList[urlChatID].ServeHTTP(writer, request)
//
//	})
//	http.HandleFunc("/chat/", func(writer http.ResponseWriter, request *http.Request) {
//		segs := strings.Split(request.URL.Path, "/")
//		urlChatID := segs[len(segs)-1]
//
//		if c.chatList[urlChatID] == nil {
//			writer.WriteHeader(http.StatusBadRequest)
//			// TODO better handle error
//			log.Fatalln("Bad URL: chat seems to not exist")
//			// panic("Bad URL")
//		}
//
//		tmpH := templateHandler{
//			filename: "chat.html",
//			data: map[string]interface{}{
//				"ClientIP": c.userIP,
//				"Host":     request.Host,
//				"ChatID":   urlChatID,
//			},
//		}
//
//		tmpH.ServeHTTP(writer, request)
//
//		//c.chatList[urlChatID].ServeHTTP(writer,request)
//
//	})
//
//	// start the web gql
//	log.Println("Starting web gql on", localServerAddress)
//	if err := http.ListenAndServe(localServerAddress, nil); err != nil {
//		log.Fatal("ListenAndServe:", err)
//	}
//}

// TestSetup setups test env
// TEST SETUP
func (c *Client) TestSetup() {
	// necessary setup
	go c.connectionsHandler()
	go c.receivedPayloadHandler()
	go c.eventListener()

	var participants []string

	tmpFriend := "tcp://127.0.0.3:7878"

	c.FriendsList[tmpFriend] = &gql.Friend{
		Nick:   &tmpFriend,
		UserID: tmpFriend,
		UserIP: &tmpFriend,
	}

	logger.Debug("TestSetup: friendslist = ", c.FriendsList)

	if value, ok := os.LookupEnv("SAMPLE_CHAT_SETUP_ADDR"); ok {
		participants = strings.Split(value, ",")
		logger.Info("TestSetup: chat setup connect to = ", participants)

		for _, part := range participants {
			tmpNick := part
			c.FriendsList[part] = &gql.Friend{
				Nick:       &tmpNick,
				UserID:     part,
				UserIP: 	&tmpNick,
			}
		}
	}

	logger.Debug("TestSetup: and now, friendslist = ", c.FriendsList)

	if value, ok := os.LookupEnv("MAIN_MACHINE"); ok && value == "0" {
		time.Sleep(5 * time.Second)
		log.Println("Main Machine")
		c.CreateChat(participants)
	} else if value, ok := os.LookupEnv("MAIN_MACHINE"); ok && value == "1"  {
		log.Println("Second Machine")
	} else if value, ok := os.LookupEnv("MAIN_MACHINE"); ok && value == "2" {
		log.Println("Third Machine")
	}

	//time.Sleep(10 * time.Minute)

}
