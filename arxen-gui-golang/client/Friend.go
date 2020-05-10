package client

// struct for friends store
type Friend struct {
	Name      string `json:"name"`
	FriendIP  string `json:"friendIP"`
	FriendID  string `json:"friendID"`
	PublicKey string `json:"publicKey"` // to be changed
}
