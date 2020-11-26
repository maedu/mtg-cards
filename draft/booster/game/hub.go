package game

type MessageType string

const (
	ReceivePlayerName    MessageType = "playerName"
	ReceiveCardSelection MessageType = "cardSelection"
	SendMessage          MessageType = "message"
	SendBooster          MessageType = "booster"
)

type Message struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	game Game
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		game:       InitGame("abc", "Commander Legends"), // TODO change
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			message := &Message{
				Type:    SendMessage,
				Content: "Welcome",
			}
			client.send <- message
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
