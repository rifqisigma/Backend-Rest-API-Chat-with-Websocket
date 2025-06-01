package ws

import (
	"api_chat_ws/dto"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	MemberId uint
	GroupID  uint
	Conn     *websocket.Conn
	Send     chan []byte
}

type BroadcastMessage struct {
	GroupID uint
	Action  string
	Message []byte
}

type Hub struct {
	Groups     map[uint]map[uint]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan BroadcastMessage
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Groups:     make(map[uint]map[uint]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan BroadcastMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Groups[client.GroupID] == nil {
				h.Groups[client.GroupID] = make(map[uint]*Client)
			}
			h.Groups[client.GroupID][client.MemberId] = client
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if groupClients, ok := h.Groups[client.GroupID]; ok {
				if _, ok := groupClients[client.MemberId]; ok {
					delete(groupClients, client.MemberId)
					close(client.Send)
				}
				if len(groupClients) == 0 {
					delete(h.Groups, client.GroupID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			groupClients := h.Groups[msg.GroupID]
			for _, client := range groupClients {
				select {
				case client.Send <- msg.Message:
				default:
					close(client.Send)
					delete(groupClients, client.MemberId)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) GetClientsByGroupID(groupID uint) []dto.MemberStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clientsMap, ok := h.Groups[groupID]
	if !ok {
		return nil
	}

	clients := make([]dto.MemberStatus, 0, len(clientsMap))
	for _, client := range clientsMap {
		clients = append(clients, dto.MemberStatus{
			MemberId: client.MemberId,
			Status:   true,
		})
	}

	return clients
}
