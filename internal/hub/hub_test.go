package hub_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"trivia/internal/hub"
)

// startTestServer creates a test HTTP server that upgrades connections to WebSocket
// and registers/deregisters clients in the given room.
// The handler runs until the WebSocket connection closes.
func startTestServer(t *testing.T, h *hub.Hub, room hub.Room) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		c := hub.NewClient(conn, room, "")
		h.Register(c)
		defer h.Deregister(c)
		// Drain incoming messages until the connection closes.
		ctx := context.Background()
		for {
			var raw interface{}
			if err := wsjson.Read(ctx, conn, &raw); err != nil {
				return // connection closed
			}
		}
	}))
	return srv
}

// dialWS dials the test server and returns the connection.
func dialWS(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	u := "ws" + srv.URL[4:] // replace http with ws
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

// collectMessage reads one JSON message from the connection within 500ms.
func collectMessage(t *testing.T, conn *websocket.Conn) (map[string]interface{}, bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	var msg map[string]interface{}
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		return nil, false
	}
	return msg, true
}

// -- Behavior 1+2: registering a client adds it to the correct room ----------

func TestHub_RegisterAddsClientToRoom(t *testing.T) {
	h := hub.NewHub()

	srv := startTestServer(t, h, hub.RoomHost)
	defer srv.Close()

	conn := dialWS(t, srv)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Give the server goroutine time to register.
	time.Sleep(20 * time.Millisecond)

	if got := h.ClientCount(hub.RoomHost); got != 1 {
		t.Errorf("RoomHost count = %d; want 1", got)
	}
	if got := h.ClientCount(hub.RoomPlay); got != 0 {
		t.Errorf("RoomPlay count = %d; want 0", got)
	}
	if got := h.ClientCount(hub.RoomDisplay); got != 0 {
		t.Errorf("RoomDisplay count = %d; want 0", got)
	}
}

// -- Behavior 3: deregistering removes the client ----------------------------

func TestHub_DeregisterRemovesClient(t *testing.T) {
	h := hub.NewHub()

	srv := startTestServer(t, h, hub.RoomPlay)
	defer srv.Close()

	conn := dialWS(t, srv)
	time.Sleep(20 * time.Millisecond)

	if got := h.ClientCount(hub.RoomPlay); got != 1 {
		t.Fatalf("after register: RoomPlay count = %d; want 1", got)
	}

	// Close the client connection — server handler returns and deregisters.
	conn.Close(websocket.StatusNormalClosure, "bye")

	// Poll up to 200ms for the server goroutine to deregister.
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if h.ClientCount(hub.RoomPlay) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("after deregister: RoomPlay count = %d; want 0", h.ClientCount(hub.RoomPlay))
}

// -- Behavior 4: broadcast reaches all clients in target room ----------------

func TestHub_BroadcastReachesAllClientsInRoom(t *testing.T) {
	h := hub.NewHub()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		c := hub.NewClient(conn, hub.RoomDisplay, "")
		h.Register(c)
		defer h.Deregister(c)
		ctx := context.Background()
		for {
			var raw interface{}
			if err := wsjson.Read(ctx, conn, &raw); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	// Connect three display clients.
	wsURL := "ws" + srv.URL[4:]
	dialCtx := context.Background()
	c1, _, _ := websocket.Dial(dialCtx, wsURL, nil)
	c2, _, _ := websocket.Dial(dialCtx, wsURL, nil)
	c3, _, _ := websocket.Dial(dialCtx, wsURL, nil)
	defer c1.Close(websocket.StatusNormalClosure, "")
	defer c2.Close(websocket.StatusNormalClosure, "")
	defer c3.Close(websocket.StatusNormalClosure, "")

	time.Sleep(30 * time.Millisecond) // wait for all registrations

	payload := map[string]interface{}{"event": "test_event", "payload": map[string]interface{}{"x": 1}}
	if err := h.Broadcast(hub.RoomDisplay, payload); err != nil {
		t.Fatalf("Broadcast error: %v", err)
	}

	for i, conn := range []*websocket.Conn{c1, c2, c3} {
		msg, ok := collectMessage(t, conn)
		if !ok {
			t.Errorf("client %d: expected message, got none", i+1)
			continue
		}
		if msg["event"] != "test_event" {
			t.Errorf("client %d: event = %q; want %q", i+1, msg["event"], "test_event")
		}
	}
}

// -- Behavior 5: broadcast does NOT reach clients in other rooms -------------

func TestHub_BroadcastDoesNotCrossRoomBoundary(t *testing.T) {
	h := hub.NewHub()

	// playServer registers clients to RoomPlay.
	playServer := startTestServer(t, h, hub.RoomPlay)
	defer playServer.Close()

	// hostServer registers clients to RoomHost.
	hostServer := startTestServer(t, h, hub.RoomHost)
	defer hostServer.Close()

	playConn := dialWS(t, playServer)
	hostConn := dialWS(t, hostServer)
	defer playConn.Close(websocket.StatusNormalClosure, "")
	defer hostConn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(30 * time.Millisecond)

	// Broadcast only to RoomHost.
	payload := map[string]interface{}{"event": "host_only", "payload": map[string]interface{}{}}
	if err := h.Broadcast(hub.RoomHost, payload); err != nil {
		t.Fatalf("Broadcast error: %v", err)
	}

	// hostConn should receive the message.
	if _, ok := collectMessage(t, hostConn); !ok {
		t.Error("host client: expected message, got none")
	}

	// playConn should NOT receive anything within 100ms.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var silenceCheck map[string]interface{}
	err := wsjson.Read(ctx, playConn, &silenceCheck)
	if err == nil {
		t.Errorf("play client received unexpected message: %v", silenceCheck)
	}
}

// -- Behavior 6: three named rooms exist -------------------------------------

func TestHub_ThreeNamedRoomsExist(t *testing.T) {
	h := hub.NewHub()

	for _, room := range []hub.Room{hub.RoomHost, hub.RoomPlay, hub.RoomDisplay} {
		if got := h.ClientCount(room); got != 0 {
			t.Errorf("room %q: initial count = %d; want 0", room, got)
		}
	}
}

// -- Behavior 6 (concurrent): concurrent register/deregister is race-free ---

func TestHub_ConcurrentRegisterDeregisterNoRace(t *testing.T) {
	h := hub.NewHub()
	srv := startTestServer(t, h, hub.RoomPlay)
	defer srv.Close()

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", srv.Listener.Addr().String())
			if err != nil {
				return
			}
			conn.Close()
		}()
	}
	wg.Wait()

	// Exercise the broadcast path concurrently too.
	var wg2 sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			_ = h.Broadcast(hub.RoomPlay, map[string]interface{}{"event": "ping", "payload": map[string]interface{}{}})
		}()
	}
	wg2.Wait()
}
