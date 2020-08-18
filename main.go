package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

var users, anchors sync.Map
var upgrader = websocket.Upgrader{}

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

var messChan = make(chan Mess)

func anchorBroadCast(anchor Anchor) {
	var mess Mess
	for {
		err := anchor.Conn.ReadJSON(&mess)
		if err != nil {
			anchor.Conn.Close()
			anchors.Delete(anchor.Key)
			return
		}
		if mess.Event != "message" {
			if user, ok := users.Load(mess.Receiver); ok {
				user.(User).Conn.WriteJSON(Mess{
					Event:    mess.Event,
					Data:     mess.Data,
					Name:     mess.Name,
					Receiver: mess.Receiver,
				})
			}
		} else {
			mess.Receiver = anchor.Key
			messChan <- mess
		}
	}
}

func userBroadCast(u User) {
	var mess Mess
	for {
		err := u.Conn.ReadJSON(&mess)
		if err != nil {
			u.Conn.Close()
			users.Delete(u.Key)
			return
		}
		if mess.Event != "message" {
			if conn, ok := anchors.Load(mess.Receiver); ok {
				conn.(Anchor).Conn.WriteJSON(Mess{
					Event:    mess.Event,
					Data:     mess.Data,
					Name:     mess.Name,
					Receiver: mess.Receiver,
				})
			}
		} else {
			messChan <- mess
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.Close()
		log.Fatal(err)
	}
	u, _ := url.Parse(r.RequestURI)
	query := u.Query()
	if len(query) == 1 {
		if _, ok := anchors.Load(query.Get("name")); !ok {
			anchor := Anchor{Conn: c, Key: query.Get("name")}
			anchors.Store(query.Get("name"), anchor)
			go anchorBroadCast(anchor)
		} else {
			c.Close()
			anchors.Delete(query.Get("name"))
			return
		}
	}
	if len(query) == 2 {
		if _, ok := users.Load(query.Get("name")); !ok {
			user := User{Conn: c, Key: query.Get("name"), Ar: query.Get("receiver")}
			users.Store(query.Get("name"), user)
			if anchorConn, ok := anchors.Load(query.Get("receiver")); ok {
				data := map[string]string{
					"name":     query.Get("name"),
					"receiver": query.Get("receiver"),
				}
				log.Println("user connect...:", user)
				err = anchorConn.(Anchor).Conn.WriteJSON(data)
				if err != nil {
					log.Fatal(err)
				}
			}
			go userBroadCast(user)
		} else {
			users.Delete(query.Get("name"))
			c.Close()
			return
		}
	}
}

func messHandler() {
	for {
		select {
		case mess := <-messChan:
			data := fmt.Sprintf("%s:%s", mess.Name, mess.Data.(string))
			users.Range(func(user, conn interface{}) bool {
				if conn.(User).Ar != mess.Receiver {
					return true
				}
				err := conn.(User).Conn.WriteJSON(Mess{
					Event: "message",
					Data:  data,
				})
				if err != nil {
					conn.(User).Conn.Close()
					users.Delete(user)
				}
				return true
			})
			if conn, ok := anchors.Load(mess.Receiver); ok {
				err := conn.(Anchor).Conn.WriteJSON(Mess{
					Event: "message",
					Data:  data,
				})
				if err != nil {
					conn.(Anchor).Conn.Close()
					anchors.Delete(mess.Receiver)
				}
			}
		}
	}
}

func main() {
	go messHandler()
	http.HandleFunc("/websocket", echo)
	http.Handle("/", http.FileServer(http.Dir("./app/public")))
	log.Println("Serving at localhost:3000...")
	log.Fatal(http.ListenAndServeTLS("0.0.0.0:3000", "./scg130.com+3.pem","./scg130.com+3-key.pem",nil))
}
