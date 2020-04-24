package client

import (
	"context"
	"encoding/json"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"html/template"
	"log"
	"main/chat"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

const CONNECTIONS_UPDATE_REFRESH_RATE = 10 * time.Second

type Client struct {
	userIP     string
	clientsIPs map[string]bool // clientIP : status
	// not in use rn
	clientsSockets      map[rsocket.Client]string       // socket : clientIP
	chatList            map[string]*chat.Chat           // chatID, *Chat
	sendDataList        map[string]chan payload.Payload // payload and target chat format: map[clientIP] payload(message, chatID)
	receivedPayloadChan chan payload.Payload            // channel with all incoming payloads

	friendsList map[string]Friend // map[friendsNick]Friend
	secretKey   string            // used for authentication
}

// return new Client
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

	return &Client{
		userIP:              userAddr,
		clientsIPs:          _clientsIPs,
		clientsSockets:      _clientsSockets,
		chatList:            _chatList,
		sendDataList:        _sendMessageList,
		receivedPayloadChan: _receivedPayloadChan,
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

	// TODO add chat ID generator
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

	// advert new chat
	c.receivedPayloadChan <- payload.New([]byte(chatIDstr), c.getMetadataTag(CHAT_ADVERT_REQUEST))

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

// handler of all connections across itself and other clients
func (c *Client) connectionsHandler() {
	for {
		// refresh at rate
		time.Sleep(CONNECTIONS_UPDATE_REFRESH_RATE)

		for addr, status := range c.clientsIPs {
			// if client not connected to particular client try to connect
			if !status {
				// find if chan for that client exists
				// TODO can be written better
				if c.sendDataList[addr] == nil {
					log.Println("connectionsHandler: chan non existing - creating ", addr)
					ch := make(chan payload.Payload)
					c.sendDataList[addr] = ch
				}
				go c.connectToClient(c.sendDataList[addr], addr)
				// after finished update record
				c.clientsIPs[addr] = true
			}
		}
	}
}

// Possible type problem: struct vs payload
func (c *Client) connectToClient(ch chan payload.Payload, addr string) {
	// goroutine for connecting to clients
	// handle channels

	// in advanced scenario ask host for chat clients ips

	// create tmp flux
	// TODO problem: who is the target
	// TODO add option of sending custom messages
	f := flux.Create(func(ctx context.Context, s flux.Sink) {
		log.Println("STARTED sending new message")
		for mess := range ch {
			log.Println("SENDING new message")
			s.Next(mess)
		}
		s.Complete()
	})

	// new client
	// TODO change literals to constants
	cli, err := rsocket.
		Connect().
		SetupPayload(payload.NewString(c.userIP, "1234")).
		Resume().
		Fragment(1024).
		Transport(addr).
		Start(context.Background())
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	log.Println("REQUESTING CHANNEL WITH ", addr)

	// possible error
	// TODO remove debug stats
	_, err = cli.RequestChannel(f).
		DoOnNext(func(elem payload.Payload) {
			log.Println("GOT new message")
			tmpChatID, _ := elem.MetadataUTF8()
			c.chatList[tmpChatID].MessagesChan <- elem
		}).
		BlockLast(context.Background())
}

// runs eventListener() and manages connections
func (c *Client) clientManager() {

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

				// TODO sort messages

				log.Println("GOT MESSAGE: ", input.DataUTF8())
				tmpChatID, _ := input.MetadataUTF8()
				c.chatList[tmpChatID].MessagesChan <- input
			}))

			return flux.Create(func(ctx context.Context, s flux.Sink) {
				for mess := range c.sendDataList[setup.DataUTF8()] {
					s.Next(mess)
				}
				s.Complete()
			})
		}),
	)
}

// payloads:
// CHAT_MESSAGE:			  {message,{source, type, chatID}}
// CHAT_PARTICIPANTS_REQUEST: {chatID, {source, type}}

// helper, handling all incoming messages from each connection
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
		switch metadata["type"].(string) {
		case CHAT_MESSAGE:
			// TODO handle incoming messages
		case CHAT_PARTICIPANTS_REQUEST:
			// send all participating clients IPs to requester
			for _, addr := range c.chatList[payl.DataUTF8()].ClientsIPsList() {
				tmpSendPayl := payload.New([]byte(addr), c.getMetadataTag(CHAT_PARTICIPANTS_RESPONSE))
				c.sendDataList[metadata["source"].(string)] <- tmpSendPayl
			}
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
					c.sendDataList[addr] <- payload.New(payl.Data(), c.getMetadataTag(CHAT_ADVERT))
				}
			}
		case CHAT_ADVERT:
			// TODO implement me
		}

	}
}

// function returning metadata for payload
// args:
// args[0]: type of request/response to be generated
// args[1:]: extra arguments:
func (c *Client) getMetadataTag(args ...string) []byte {
	switch args[0] {
	case CHAT_PARTICIPANTS_RESPONSE:
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `"}`)
	case CHAT_MESSAGE:
		// args[1]: chatID
		if len(args) < 2 {
			panic("getMetadataTag: Too few arguments")
		}
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `", "chatID":"` + args[1] + `"}`)
	case CHAT_ADVERT_REQUEST:
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `"}`)
	case CHAT_ADVERT:
		return []byte(`{"source":"` + c.userIP + `", "type":"` + args[0] + `"}`)
	default:
		log.Fatalln("getMetadataTag: Bad message type")
		return nil
	}
}

// --------------------------------------------------------
// web part
// --------------------------------------------------------

const localServerAddress = "127.0.0.1:7879"

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
	data     map[string]interface{}
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	//data := map[string]interface{}{
	//	"Host": r.Host,
	//}

	t.templ.Execute(w, t.data)
}

func (c *Client) HttpServer() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// TODO implement me
	})
	http.Handle("/fixme", &templateHandler{filename: "chat.html"})
	http.HandleFunc("/room/", func(writer http.ResponseWriter, request *http.Request) {
		segs := strings.Split(request.URL.Path, "/")
		urlChatID := segs[len(segs)-1]

		if c.chatList[urlChatID] == nil {
			writer.WriteHeader(http.StatusBadRequest)
			// TODO better handle error
			log.Fatalln("Bad URL: chat seems to not exist")
			// panic("Bad URL")
		}

		c.chatList[urlChatID].ServeHTTP(writer, request)

	})
	http.HandleFunc("/chat/", func(writer http.ResponseWriter, request *http.Request) {
		segs := strings.Split(request.URL.Path, "/")
		urlChatID := segs[len(segs)-1]

		if c.chatList[urlChatID] == nil {
			writer.WriteHeader(http.StatusBadRequest)
			// TODO better handle error
			log.Fatalln("Bad URL: chat seems to not exist")
			// panic("Bad URL")
		}

		tmpH := templateHandler{
			filename: "templates/chat.html",
			data: map[string]interface{}{
				"Host":   request.Host,
				"ChatID": urlChatID,
			},
		}

		tmpH.ServeHTTP(writer, request)

		//c.chatList[urlChatID].ServeHTTP(writer,request)

	})

	// start the web server
	log.Println("Starting web server on", localServerAddress)
	if err := http.ListenAndServe(localServerAddress, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// TEST SETUP

func (c *Client) TestSetup() {
	// necessary setup
	go c.connectionsHandler()
	go c.receivedPayloadHandler()
	go c.eventListener()

	participants := []string{"tcp://127.0.0.3:7878"}

	if value, ok := os.LookupEnv("SAMPLE_CHAT_SETUP_ADDR"); ok {
		participants = strings.Split(value, ",")
		log.Println("TestSetup: chat setup/ connect to = " + participants[0])
	}

	if value, ok := os.LookupEnv("MAIN_MACHINE"); ok && value == "1" {
		time.Sleep(5*time.Second)
		log.Println("Main Machine")
		c.createChat(participants)
	} else {
		log.Println("Second Machine")
	}

	time.Sleep(3*time.Minute)

}
