package chat

type Hub struct {
	clients    map[*Client]struct{}
	register   chan registerRequest
	unregister chan *Client
	broadcast  chan []byte
	maxClients int
	stopCh     chan struct{}
}

type registerRequest struct {
	client   *Client
	accepted chan bool
}

func NewHub(maxClients int) *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan registerRequest),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		maxClients: maxClients,
		stopCh:     make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.stopCh:
			h.closeAllClients()
			return
		case req := <-h.register:
			client := req.client
			if len(h.clients) >= h.maxClients {
				req.accepted <- false
				continue
			}
			h.clients[client] = struct{}{}
			req.accepted <- true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (h *Hub) Close() { close(h.stopCh) }

func (h *Hub) closeAllClients() {
	for client := range h.clients {
		close(client.send)
	}
}
