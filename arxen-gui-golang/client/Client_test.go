package client

import (
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"main/chat"
	"net"
	"reflect"
	"testing"
)

func TestClient_recivedPayloadHandler(t *testing.T) {
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
		name   string
		source string
		fields fields
	}{
		{"test_CHAT_PARTICIPANTS_REQUEST", "tcp://10.5.0.2:7878",
			fields{
				userIP:              "tcp://10.5.0.3:7878",
				receivedPayloadChan: make(chan payload.Payload),
				sendDataList:        make(map[string]chan payload.Payload),
				clientsIPs:          make(map[string]bool),
				chatList:            make(map[string]*chat.Chat),
			}},
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

			c.createChat([]string{"1", "2", "3", "4"})

			ch := make(chan payload.Payload)
			c.sendDataList[tt.source] = ch

			data01 := payload.New([]byte(`123`), []byte(`{"source":"`+tt.source+`", "type":"CHAT_PARTICIPANTS_REQUEST"}`))

			go c.receivedPayloadHandler()

			c.receivedPayloadChan <- data01

			var rcvData02 []payload.Payload

			for data := range c.sendDataList[tt.source] {
				rcvData02 = append(rcvData02, data)
				if len(rcvData02) == 3 {
					break
				}
			}

			resp := payload.New([]byte(`2`), []byte(`{"source":"`+c.userIP+`", "type":"CHAT_PARTICIPANTS_RESPONSE"}`))

			if rcvData02[1].DataUTF8() != resp.DataUTF8() {
				t.Errorf("Test failed: \"%v\" is not equal to \"%v\"", rcvData02[1], payload.New([]byte("2"),
					[]byte(`{"source":"`+c.userIP+`", "type":"CHAT_PARTICIPANTS_RESPONSE"}`)))
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
			if got := NewUser(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
