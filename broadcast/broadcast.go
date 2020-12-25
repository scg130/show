package broadcast

import "show/dto"

func UserBroadCast(u dto.User) {
	var mess dto.Mess
	for {
		err := u.Conn.ReadJSON(&mess)
		if err != nil {
			u.Conn.Close()
			dto.Users.Delete(u.Key)
			return
		}
		u.Conn.SetReadDeadline(dto.Expire)
		if mess.Data == nil {
			continue
		}
		if mess.Event != "message" {
			if conn, ok := dto.Anchors.Load(mess.Receiver); ok {
				if err := conn.(dto.Anchor).Conn.WriteJSON(dto.Mess{
					Event:    mess.Event,
					Data:     mess.Data,
					Name:     mess.Name,
					Receiver: mess.Receiver,
				}); err != nil {
					conn.(dto.Anchor).Conn.Close()
					dto.Anchors.Delete(conn.(dto.Anchor).Key)
				}
			}
		} else if mess.Event == "ping" {
			continue
		} else {
			dto.MessChan <- mess
		}
	}
}

func AnchorBroadCast(anchor dto.Anchor) {
	var mess dto.Mess
	for {
		err := anchor.Conn.ReadJSON(&mess)
		if err != nil {
			anchor.Conn.Close()
			dto.Anchors.Delete(anchor.Key)
			return
		}
		anchor.Conn.SetReadDeadline(dto.Expire)
		if mess.Data == nil {
			continue
		}
		if mess.Event != "message" {
			if user, ok := dto.Users.Load(mess.Receiver); ok {
				user.(dto.User).Conn.WriteJSON(dto.Mess{
					Event:    mess.Event,
					Data:     mess.Data,
					Name:     mess.Name,
					Receiver: mess.Receiver,
				})
			}
		} else if mess.Event == "ping" {
			continue
		} else {
			mess.Receiver = anchor.Key
			dto.MessChan <- mess
		}
	}
}