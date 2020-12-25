package dto

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var Expire = time.Now().Add(time.Minute * 3)
var Users, Anchors sync.Map
var Upgrader = websocket.Upgrader{}

type User struct {
	Key  string
	Conn *websocket.Conn
	Ar   string
}

type Anchor struct {
	Conn *websocket.Conn
	Key  string
}

type Mess struct {
	Event    string      `json:"event"`
	Data     interface{} `json:"data"`
	Name     string      `json:"name"`
	Receiver string      `json:"receiver"`
}

var MessChan = make(chan Mess)

