package realtime

import (
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type mockSender struct {
	ch chan OutgoingMessage
}

func (m *mockSender) Send() chan OutgoingMessage {
	return m.ch
}

type testHarness struct {
	hub      *hub
	clients  []*mockSender
}

func newTestHub(t *testing.T, numClients int) *testHarness {
	t.Helper()

	h := newHub(slog.Default(), "test-room")
	go h.Run()

	clients := make([]*mockSender, numClients)
	for i := 0; i < numClients; i++ {
		c := &mockSender{ch: make(chan OutgoingMessage, 64)}
		h.Register() <- c
		// drain sync message
		<-c.ch
		clients[i] = c
	}

	return &testHarness{hub: h, clients: clients}
}

func (th *testHarness) close() {
	th.hub.Close()
}

func (th *testHarness) recv(idx int) OutgoingMessage {
	select {
	case msg := <-th.clients[idx].ch:
		return msg
	case <-time.After(time.Second):
		panic("timeout waiting for message")
	}
}

func (th *testHarness) recvAll() []OutgoingMessage {
	var msgs []OutgoingMessage
	for i := range th.clients {
		msgs = append(msgs, th.recv(i))
	}
	return msgs
}

func (th *testHarness) recvNone(idx int) bool {
	select {
	case <-th.clients[idx].ch:
		return false
	default:
		return true
	}
}

func TestHubRegister_SendsSync(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	go h.Run()
	defer h.Close()

	client := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- client

	msg := <-client.ch
	if msg.Type != "sync" {
		t.Errorf("expected sync, got %s", msg.Type)
	}

	payload, ok := msg.Payload.(SyncPayload)
	if !ok {
		t.Fatalf("expected SyncPayload, got %T", msg.Payload)
	}
	if payload.Playing {
		t.Error("expected Playing=false")
	}
}

func TestHubRegister_SyncIncludesCurrentState(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	h.state.SourceID = "s1"
	h.state.SourceURL = "https://example.com"
	h.state.Playing = true
	h.state.Position = 42.5

	go h.Run()
	defer h.Close()

	client := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- client

	msg := <-client.ch
	payload := msg.Payload.(SyncPayload)

	if payload.SourceID != "s1" {
		t.Errorf("expected SourceID=s1, got %s", payload.SourceID)
	}
	if payload.SourceURL != "https://example.com" {
		t.Errorf("expected SourceURL=https://example.com, got %s", payload.SourceURL)
	}
	if !payload.Playing {
		t.Error("expected Playing=true")
	}
	if payload.Position != 42.5 {
		t.Errorf("expected Position=42.5, got %f", payload.Position)
	}
}

func TestHubUnregister_ShutsDownOnEmpty(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	done := make(chan struct{})
	go func() {
		h.Run()
		close(done)
	}()

	client := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- client
	<-client.ch // drain sync

	h.Unregister() <- client

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("hub did not shut down after last client unregistered")
	}
}

func TestHubUnregister_RemainingClientsKeepHubAlive(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	go h.Run()
	defer h.Close()

	c1 := &mockSender{ch: make(chan OutgoingMessage, 10)}
	c2 := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- c1
	<-c1.ch
	h.Register() <- c2
	<-c2.ch

	h.Unregister() <- c1

	// hub should still be alive — send using a sender that is not a registered client
	payload, _ := json.Marshal(map[string]string{"text": "hi"})
	h.Incoming() <- incomingEvent{
		Username: "alice",
		Message:  IncomingMessage{Type: "chat", Payload: payload},
	}

	msg := <-c2.ch
	if msg.Type != "chat" {
		t.Errorf("expected chat, got %s", msg.Type)
	}
}

func TestHubStop_ClosesAllClients(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	go h.Run()

	c1 := &mockSender{ch: make(chan OutgoingMessage, 10)}
	c2 := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- c1
	<-c1.ch
	h.Register() <- c2
	<-c2.ch

	h.Close()

	_, ok1 := <-c1.ch
	if ok1 {
		t.Error("expected c1 channel to be closed")
	}
	_, ok2 := <-c2.ch
	if ok2 {
		t.Error("expected c2 channel to be closed")
	}
}

func TestHandleEvent_Chat_BroadcastsToAllExceptSender(t *testing.T) {
	th := newTestHub(t, 3)

	payload, _ := json.Marshal(map[string]string{"text": "hello"})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		UserID:   "u1",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "chat", Payload: payload},
	}

	// sender (client 0) should NOT get the message
	if !th.recvNone(0) {
		t.Error("sender should not receive own chat message")
	}

	// all others should get it
	for i := 1; i < 3; i++ {
		msg := th.recv(i)
		if msg.Type != "chat" {
			t.Errorf("client %d: expected chat, got %s", i, msg.Type)
		}
		if msg.Username != "alice" {
			t.Errorf("client %d: expected Username=alice, got %s", i, msg.Username)
		}
	}
}

func TestHandleEvent_Chat_EmptyTextIgnored(t *testing.T) {
	th := newTestHub(t, 2)

	payload, _ := json.Marshal(map[string]string{"text": "  "})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "chat", Payload: payload},
	}

	if !th.recvNone(1) {
		t.Error("empty chat message should not be broadcast")
	}
}

func TestHandleEvent_Chat_TrimsAndTruncates(t *testing.T) {
	th := newTestHub(t, 2)

	longText := strings.Repeat("a", maxTextLen+50)
	payload, _ := json.Marshal(map[string]string{"text": longText})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "chat", Payload: payload},
	}

	msg := th.recv(1)
	p := msg.Payload.(ChatPayload)
	if len(p.Text) > maxTextLen {
		t.Errorf("expected truncated text <= %d, got %d", maxTextLen, len(p.Text))
	}
	if p.Text != strings.Repeat("a", maxTextLen) {
		t.Error("expected exactly maxTextLen characters")
	}
}

func TestHandleEvent_Chat_InvalidPayloadIgnored(t *testing.T) {
	th := newTestHub(t, 2)

	payload, _ := json.Marshal(map[string]int{"wrong": 1})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "chat", Payload: payload},
	}

	if !th.recvNone(1) {
		t.Error("invalid chat payload should be ignored")
	}
}

func TestHandleEvent_Play_UpdatesStateAndBroadcasts(t *testing.T) {
	th := newTestHub(t, 2)

	payload, _ := json.Marshal(PlayerPayload{Position: 42.5})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "play", Payload: payload},
	}

	// broadcast to all (including sender)
	for i := 0; i < 2; i++ {
		msg := th.recv(i)
		if msg.Type != "play" {
			t.Errorf("client %d: expected play, got %s", i, msg.Type)
		}
		if msg.Username != "alice" {
			t.Errorf("client %d: expected username=alice", i)
		}
	}

	st := th.hub.GetState()
	if !st.Playing {
		t.Error("expected Playing=true")
	}
	if st.Position != 42.5 {
		t.Errorf("expected Position=42.5, got %f", st.Position)
	}
	if st.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestHandleEvent_Pause_UpdatesStateAndBroadcasts(t *testing.T) {
	th := newTestHub(t, 2)

	// first play
	p1, _ := json.Marshal(PlayerPayload{Position: 100})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "play", Payload: p1},
	}
	th.recvAll()

	// then pause
	p2, _ := json.Marshal(PlayerPayload{Position: 120})
	th.hub.Incoming() <- incomingEvent{
		Username: "bob",
		Sender:   th.clients[1],
		Message:  IncomingMessage{Type: "pause", Payload: p2},
	}
	th.recvAll()

	st := th.hub.GetState()
	if st.Playing {
		t.Error("expected Playing=false after pause")
	}
	if st.Position != 120 {
		t.Errorf("expected Position=120, got %f", st.Position)
	}
}

func TestHandleEvent_Play_InvalidPayloadIgnored(t *testing.T) {
	th := newTestHub(t, 1)

	payload, _ := json.Marshal(map[string]string{"bad": "data"})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "play", Payload: payload},
	}

	if !th.recvNone(0) {
		t.Error("invalid play payload should not broadcast")
	}

	st := th.hub.GetState()
	if st.Playing {
		t.Error("state should not change on invalid payload")
	}
}

func TestHandleEvent_Seek_UpdatesPositionAndBroadcasts(t *testing.T) {
	th := newTestHub(t, 2)

	// first play
	p1, _ := json.Marshal(PlayerPayload{Position: 50})
	th.hub.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   th.clients[0],
		Message:  IncomingMessage{Type: "play", Payload: p1},
	}
	th.recvAll()

	// seek
	p2, _ := json.Marshal(PlayerPayload{Position: 200})
	th.hub.Incoming() <- incomingEvent{
		Username: "bob",
		Sender:   th.clients[1],
		Message:  IncomingMessage{Type: "seek", Payload: p2},
	}
	th.recvAll()

	st := th.hub.GetState()
	if !st.Playing {
		t.Error("seek should not change playing state")
	}
	if st.Position != 200 {
		t.Errorf("expected Position=200, got %f", st.Position)
	}
}

func TestHandleEvent_SourceChanged_UpdatesStateAndBroadcasts(t *testing.T) {
	th := newTestHub(t, 2)

	th.hub.BroadcastSourceChanged("s1", "https://example.com")

	for i := 0; i < 2; i++ {
		msg := th.recv(i)
		if msg.Type != "source_changed" {
			t.Errorf("client %d: expected source_changed, got %s", i, msg.Type)
		}
	}

	st := th.hub.GetState()
	if st.SourceID != "s1" {
		t.Errorf("expected SourceID=s1, got %s", st.SourceID)
	}
	if st.SourceURL != "https://example.com" {
		t.Errorf("expected SourceURL=https://example.com, got %s", st.SourceURL)
	}
}

func TestHandleEvent_SourceChanged_OverwritesPrevious(t *testing.T) {
	th := newTestHub(t, 1)

	th.hub.BroadcastSourceChanged("s1", "https://one.com")
	th.recv(0)

	th.hub.BroadcastSourceChanged("s2", "https://two.com")
	th.recv(0)

	st := th.hub.GetState()
	if st.SourceID != "s2" {
		t.Errorf("expected SourceID=s2, got %s", st.SourceID)
	}
}

func TestHandleEvent_SourceChanged_NilSender(t *testing.T) {
	th := newTestHub(t, 1)

	th.hub.BroadcastSourceChanged("s1", "https://example.com")

	msg := th.recv(0)
	if msg.Type != "source_changed" {
		t.Errorf("expected source_changed, got %s", msg.Type)
	}
}

func TestBroadcastAll_SendsToAllClients(t *testing.T) {
	th := newTestHub(t, 3)

	msg := OutgoingMessage{Type: "test", Payload: "data"}
	th.hub.broadcastAll(msg)

	for i := 0; i < 3; i++ {
		got := th.recv(i)
		if got.Type != "test" {
			t.Errorf("client %d: expected test, got %s", i, got.Type)
		}
	}
}

func TestBroadcastAll_RemovesSlowClient(t *testing.T) {
	smallBuf := make(chan OutgoingMessage, 1)
	th := &testHarness{
		hub: newHub(slog.Default(), "room1"),
		clients: []*mockSender{
			{ch: smallBuf},
		},
	}

	go th.hub.Run()

	th.hub.Register() <- th.clients[0]
	<-th.clients[0].ch // drain sync

	// fill the buffer
	th.clients[0].ch <- OutgoingMessage{Type: "fill"}

	// this broadcast should hit default case and remove the client
	th.hub.broadcastAll(OutgoingMessage{Type: "overflow"})

	// drain the buffered "fill" message first
	<-th.clients[0].ch
	// now the channel should be closed
	if _, ok := <-th.clients[0].ch; ok {
		t.Error("expected slow client channel to be closed")
	}
}

func TestHubManager_GetOrCreate_CreatesAndReuses(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	h1 := m.GetOrCreate("room1")
	h2 := m.GetOrCreate("room1")
	h3 := m.GetOrCreate("room2")

	if h1 != h2 {
		t.Error("GetOrCreate should return same hub for same roomID")
	}
	if h1 == h3 {
		t.Error("GetOrCreate should return different hub for different roomID")
	}

	m.Remove("room1")
	m.Remove("room2")
}

func TestHubManager_Remove_StopsHub(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	h := m.GetOrCreate("room1")
	m.Remove("room1")

	// hub should be removed from internal map
	if _, ok := m.Get("room1"); ok {
		t.Error("hub should not exist after Remove")
	}

	// hub's Run() should have exited
	select {
	case <-h.stopCh:
	default:
		t.Error("hub stopCh should be closed after Remove")
	}
}

func TestHubManager_GetRoomState_ReturnsCurrentState(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	h := m.GetOrCreate("room1")

	client := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- client
	<-client.ch

	payload, _ := json.Marshal(PlayerPayload{Position: 99.9})
	h.Incoming() <- incomingEvent{
		Username: "alice",
		Sender:   client,
		Message:  IncomingMessage{Type: "play", Payload: payload},
	}
	<-client.ch // drain play broadcast

	st := m.GetRoomState("room1")
	if !st.Playing {
		t.Error("expected Playing=true")
	}
	if st.Position != 99.9 {
		t.Errorf("expected Position=99.9, got %f", st.Position)
	}

	m.Remove("room1")
}

func TestHubManager_GetRoomState_NonExistentReturnsEmpty(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	st := m.GetRoomState("nonexistent")
	if st.SourceID != "" || st.Playing {
		t.Error("expected zero state for non-existent room")
	}
}

func TestHubManager_GetMembersOnline_ReturnsCount(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	h := m.GetOrCreate("room1")

	c1 := &mockSender{ch: make(chan OutgoingMessage, 10)}
	c2 := &mockSender{ch: make(chan OutgoingMessage, 10)}

	h.Register() <- c1
	<-c1.ch
	h.Register() <- c2
	<-c2.ch

	if got := m.GetMembersOnline("room1"); got != 2 {
		t.Errorf("expected 2 members online, got %d", got)
	}

	h.Unregister() <- c1

	// wait for unregister to complete
	<-c1.ch

	if got := m.GetMembersOnline("room1"); got != 1 {
		t.Errorf("expected 1 member online after unregister, got %d", got)
	}

	h.Unregister() <- c2
}

func TestHubManager_BroadcastSourceChanged_Propagates(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	h := m.GetOrCreate("room1")
	client := &mockSender{ch: make(chan OutgoingMessage, 10)}
	h.Register() <- client
	<-client.ch

	m.BroadcastSourceChanged("room1", "s1", "https://example.com")

	msg := <-client.ch
	if msg.Type != "source_changed" {
		t.Errorf("expected source_changed, got %s", msg.Type)
	}

	m.Remove("room1")
}

func TestHubManager_BroadcastSourceChanged_NonExistentNoOp(t *testing.T) {
	m := NewHubManager(slog.Default()).(*hubManager)

	m.BroadcastSourceChanged("nonexistent", "s1", "https://example.com")
}

func TestGetSetState(t *testing.T) {
	h := newHub(slog.Default(), "room1")

	h.SetState("s1", "https://example.com", true, 42.5)

	st := h.GetState()
	if st.SourceID != "s1" || st.SourceURL != "https://example.com" || !st.Playing || st.Position != 42.5 {
		t.Error("GetState/SetState round trip failed")
	}
}

func TestMemberCount(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	go h.Run()
	defer h.Close()

	c1 := &mockSender{ch: make(chan OutgoingMessage, 10)}
	c2 := &mockSender{ch: make(chan OutgoingMessage, 10)}

	if n := h.MemberCount(); n != 0 {
		t.Errorf("expected 0 members, got %d", n)
	}

	h.Register() <- c1
	<-c1.ch
	if n := h.MemberCount(); n != 1 {
		t.Errorf("expected 1 member, got %d", n)
	}

	h.Register() <- c2
	<-c2.ch
	if n := h.MemberCount(); n != 2 {
		t.Errorf("expected 2 members, got %d", n)
	}
}

func TestNewHubManager(t *testing.T) {
	m := NewHubManager(slog.Default())
	if m == nil {
		t.Fatal("NewHubManager returned nil")
	}
}

func TestIncomingAccessor(t *testing.T) {
	h := newHub(slog.Default(), "room1")
	ch := h.Incoming()
	if ch == nil {
		t.Error("Incoming() returned nil")
	}
}
