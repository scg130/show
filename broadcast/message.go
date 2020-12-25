package broadcast

import (
	"fmt"
	"show/dto"
)

func MessHandler() {
	for {
		select {
		case mess := <-dto.MessChan:
			data := fmt.Sprintf("%s:%s", mess.Name, mess.Data.(string))
			dto.Users.Range(func(user, conn interface{}) bool {
				if conn.(dto.User).Ar != mess.Receiver {
					return true
				}
				err := conn.(dto.User).Conn.WriteJSON(dto.Mess{
					Event: "message",
					Data:  data,
				})
				if err != nil {
					conn.(dto.User).Conn.Close()
					dto.Users.Delete(user)
				}
				return true
			})
			if conn, ok := dto.Anchors.Load(mess.Receiver); ok {
				err := conn.(dto.Anchor).Conn.WriteJSON(dto.Mess{
					Event: "message",
					Data:  data,
				})
				if err != nil {
					conn.(dto.Anchor).Conn.Close()
					dto.Anchors.Delete(mess.Receiver)
				}
			}
		}
	}
}