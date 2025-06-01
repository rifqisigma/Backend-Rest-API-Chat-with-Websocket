package ws

import (
	"api_chat_ws/dto"
	"api_chat_ws/internal/usecase"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (c *Client) ReadPump(hub *Hub, usecase usecase.ChatUsecase) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	chats, err := usecase.LoadGroupChat(c.GroupID)
	if err == nil {
		c.Send <- chats
		_ = usecase.UpdateStatusChat(c.MemberId)
	}

	members, err := usecase.GetMembers(c.GroupID)
	if err == nil {
		membersOnline := hub.GetClientsByGroupID(c.GroupID)

		membersStatus := make([]dto.MemberStatus, 0, len(members))
		onlineMap := make(map[uint]bool)

		for _, m := range membersOnline {
			onlineMap[m.MemberId] = true
		}

		for _, member := range members {
			isOnline := onlineMap[member] // akan false jika tidak ada
			membersStatus = append(membersStatus, dto.MemberStatus{
				MemberId: member,
				Status:   isOnline,
			})
		}

		for {

			_, msg, err := c.Conn.ReadMessage()
			if err != nil {
				break
			}

			type IncomingMessage struct {
				Action  string `json:"action"`
				Content string `json:"content"`
				ID      uint   `json:"id"`
			}

			var incoming IncomingMessage
			if err := json.Unmarshal(msg, &incoming); err != nil {
				fmt.Print("error json")
				continue
			}
			switch incoming.Action {
			case "create":
				response, err := usecase.CreateChat(c.MemberId, c.GroupID, incoming.Content, membersStatus)
				if err != nil {
					continue
				}

				hub.Broadcast <- BroadcastMessage{
					GroupID: c.GroupID,
					Message: response,
				}
			case "update":
				response, err := usecase.UpdateChat(c.MemberId, c.GroupID, incoming.Content)
				if err != nil {
					continue
				}

				hub.Broadcast <- BroadcastMessage{
					GroupID: c.GroupID,
					Message: response,
				}
			case "delete":
				response, err := usecase.DeleteChat(c.MemberId, incoming.ID)
				if err != nil {
					continue
				}

				hub.Broadcast <- BroadcastMessage{
					GroupID: c.GroupID,
					Message: response,
				}
			}

		}

	}

}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Flush all queued messages
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
