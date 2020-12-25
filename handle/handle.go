package handle

import (
	"log"
	"net/http"
	"net/url"
	"show/broadcast"
	"show/dto"
)

func Echo(w http.ResponseWriter, r *http.Request) {
	c, err := dto.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.Close()
		log.Fatal(err)
	}
	u, _ := url.Parse(r.RequestURI)
	query := u.Query()
	if len(query) == 1 {
		if _, ok := dto.Anchors.Load(query.Get("name")); !ok {
			anchor := dto.Anchor{Conn: c, Key: query.Get("name")}
			dto.Anchors.Store(query.Get("name"), anchor)
			go broadcast.AnchorBroadCast(anchor)
		} else {
			c.Close()
			dto.Anchors.Delete(query.Get("name"))
			return
		}
	}
	if len(query) == 2 {
		if _, ok := dto.Users.Load(query.Get("name")); !ok {
			user := dto.User{Conn: c, Key: query.Get("name"), Ar: query.Get("receiver")}
			dto.Users.Store(query.Get("name"), user)
			if anchorConn, ok := dto.Anchors.Load(query.Get("receiver")); ok {
				data := map[string]string{
					"name":     query.Get("name"),
					"receiver": query.Get("receiver"),
				}
				log.Println("user connect...:", user)
				err = anchorConn.(dto.Anchor).Conn.WriteJSON(data)
				if err != nil {
					log.Fatal(err)
				}
				go broadcast.UserBroadCast(user)
			}

		} else {
			dto.Users.Delete(query.Get("name"))
			c.Close()
			return
		}
	}
}