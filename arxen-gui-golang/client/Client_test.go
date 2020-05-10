package client

import (
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"main/chat"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestClient_receivedPayloadHandler(t *testing.T) {
	type fields struct {
		userIP              string
		clientsIPs          map[string]bool
		clientsSockets      map[rsocket.Client]string
		chatList            map[string]*chat.Chat
		sendDataList        map[string]chan payload.Payload
		receivedPayloadChan chan payload.Payload
		secretKey           string
	}
	tests := []struct {
		name     string
		source   string
		fields   fields
		initList []string
	}{
		{"test_CHAT_PARTICIPANTS_REQUEST",
			"2",
			fields{
				userIP:              "tcp://10.5.0.3:7878",
				receivedPayloadChan: make(chan payload.Payload),
				sendDataList:        make(map[string]chan payload.Payload),
				clientsIPs:          make(map[string]bool),
				chatList:            make(map[string]*chat.Chat),
			},
			[]string{"1", "2", "3", "4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				userIP:              tt.fields.userIP,
				clientsIPs:          tt.fields.clientsIPs,
				clientsSockets:      tt.fields.clientsSockets,
				chatList:            tt.fields.chatList,
				sendDataList:        tt.fields.sendDataList,
				receivedPayloadChan: tt.fields.receivedPayloadChan,
				secretKey:           tt.fields.secretKey,
			}

			go c.receivedPayloadHandler()

			// tmp solution
			for _, addr := range tt.initList {
				ch := make(chan payload.Payload, 5)
				c.sendDataList[addr] = ch
			}

			go c.CreateChat(tt.initList)

			time.Sleep(10 * time.Millisecond)

			nameString := "123"

			for name, _ := range c.chatList {
				nameString = name
			}

			data01 := payload.New([]byte(nameString), []byte(`{"source":"`+tt.source+`", "type":"CHAT_PARTICIPANTS_REQUEST"}`))

			c.receivedPayloadChan <- data01

			var rcvData02 []payload.Payload

			quit := make(chan struct{})
			go func() {
				for {
					select {
					case <-quit:
						return
					default:
						for _, addr := range tt.initList {
							if addr != tt.source {
								<-c.sendDataList[addr]
							}
						}
					}
				}
			}()

			for data := range c.sendDataList[tt.source] {
				rcvData02 = append(rcvData02, data)
				if len(rcvData02) == 2 {
					break
				}
			}

			resp := payload.New([]byte(`1,2,3,4,tcp://10.5.0.3:7878`), []byte(`{"source":"`+c.userIP+`", "type":"CHAT_PARTICIPANTS_RESPONSE"}`))

			if rcvData02[1].DataUTF8() != resp.DataUTF8() {
				t.Errorf("Test failed: \"%v\" is not equal to \"%v\"", rcvData02[1], resp)
			}

			close(quit)
		})
	}
}

// sometimes fails due to: "fatal error: concurrent map writes"
func TestClient_CHAT_ADVERT(t *testing.T) {
	type fields struct {
		userIP              string
		clientsIPs          map[string]bool
		clientsSockets      map[rsocket.Client]string
		chatList            map[string]*chat.Chat
		sendDataList        map[string]chan payload.Payload
		receivedPayloadChan chan payload.Payload
		secretKey           string
	}
	tests := []struct {
		name            string
		otherClientsIPs []string
		fields          fields
		output          map[string][]payload.Payload
		chatID          string
	}{
		{
			"test_CHAT_ADVERT",
			[]string{"tcp://10.5.0.2:7878", "tcp://10.5.0.3:7878", "tcp://10.5.0.4:7878"},
			fields{
				userIP:              "tcp://10.5.0.1:7878",
				receivedPayloadChan: make(chan payload.Payload),
				sendDataList:        make(map[string]chan payload.Payload),
				clientsIPs:          make(map[string]bool),
				chatList:            make(map[string]*chat.Chat),
			},
			make(map[string][]payload.Payload),
			"123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				userIP:              tt.fields.userIP,
				clientsIPs:          tt.fields.clientsIPs,
				clientsSockets:      tt.fields.clientsSockets,
				chatList:            tt.fields.chatList,
				sendDataList:        tt.fields.sendDataList,
				receivedPayloadChan: tt.fields.receivedPayloadChan,
				secretKey:           tt.fields.secretKey,
			}

			var wg sync.WaitGroup

			for _, item := range tt.otherClientsIPs {
				ch := make(chan payload.Payload, 5)
				c.sendDataList[item] = ch
			}

			go c.receivedPayloadHandler()

			time.Sleep(50 * time.Millisecond)

			var mu sync.Mutex

			for _, listener := range tt.otherClientsIPs {
				wg.Add(1)
				go func(_wg *sync.WaitGroup, lis string) {
					defer _wg.Done()
					for payl := range c.sendDataList[lis] {
						mu.Lock()
						tt.output[lis] = append(tt.output[lis], payl)
						mu.Unlock()
						if len(tt.output[lis]) > 0 {
							break
						}
					}
				}(&wg, listener)
			}

			time.Sleep(50 * time.Millisecond)

			c.CreateChat(tt.otherClientsIPs)

			nameString := "123"

			for name, _ := range c.chatList {
				nameString = name
			}

			wg.Wait()

			for _, item := range tt.output {
				if item[0].DataUTF8() != nameString {
					t.Errorf("Test FAILED: output \"%s\" != \"%s\"!", item[0].DataUTF8(), nameString)
				}
			}
		})
	}
}

func TestClient_responder(t *testing.T) {
	type fields struct {
		userIP              string
		clientsIPs          map[string]bool
		clientsSockets      map[rsocket.Client]string
		chatList            map[string]*chat.Chat
		sendDataList        map[string]chan payload.Payload
		receivedPayloadChan chan payload.Payload
		secretKey           string
	}
	type args struct {
		setup payload.SetupPayload
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   rsocket.RSocket
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				userIP:              tt.fields.userIP,
				clientsIPs:          tt.fields.clientsIPs,
				clientsSockets:      tt.fields.clientsSockets,
				chatList:            tt.fields.chatList,
				sendDataList:        tt.fields.sendDataList,
				receivedPayloadChan: tt.fields.receivedPayloadChan,
				secretKey:           tt.fields.secretKey,
			}
			if got := c.responder(tt.args.setup); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("responder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOutboundIP(t *testing.T) {
	tests := []struct {
		name  string
		want  net.IP
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetOutboundIP()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOutboundIP() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetOutboundIP() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
