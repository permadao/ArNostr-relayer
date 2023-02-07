package relayer

import (
	"context"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nbd-wtf/go-nostr"
)

func TestServerStartShutdown(t *testing.T) {
	var (
		serverHost  string
		inited      bool
		storeInited bool
		shutdown    bool
	)
	ready := make(chan struct{})
	rl := &testRelay{
		name: "test server start",
		init: func() error {
			inited = true
			return nil
		},
		onInitialized: func(s *Server) {
			serverHost = s.Addr()
			close(ready)
		},
		onShutdown: func(context.Context) { shutdown = true },
		storage: &testStorage{
			init: func() error { storeInited = true; return nil },
		},
	}
	srv := NewServer("127.0.0.1:0", rl)
	done := make(chan error)
	go func() { done <- srv.Start(); close(done) }()

	// verify everything's initialized
	select {
	case <-ready:
		// continue
	case <-time.After(time.Second):
		t.Fatal("srv.Start too long to initialize")
	}
	if !inited {
		t.Error("didn't call testRelay.init")
	}
	if !storeInited {
		t.Error("didn't call testStorage.init")
	}

	// check that http requests are served
	if _, err := http.Get("http://" + serverHost); err != nil {
		t.Errorf("GET %s: %v", serverHost, err)
	}

	// verify server shuts down
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("srv.Shutdown: %v", err)
	}
	if !shutdown {
		t.Error("didn't call testRelay.onShutdown")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("srv.Start: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("srv.Start too long to return")
	}
}

func TestServerShutdownWebsocket(t *testing.T) {
	// set up a new relay server
	srv := startTestRelay(t, &testRelay{storage: &testStorage{}})

	// connect a client to it
	ctx1, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := nostr.RelayConnect(ctx1, "ws://"+srv.Addr())
	if err != nil {
		t.Fatalf("nostr.RelayConnectContext: %v", err)
	}

	// now, shut down the server
	ctx2, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx2); err != nil {
		t.Errorf("srv.Shutdown: %v", err)
	}

	// wait for the client to receive a "connection close"
	select {
	case err := <-client.ConnectionError:
		if _, ok := err.(*websocket.CloseError); !ok {
			t.Errorf("client.ConnextionError: %v (%T); want websocket.CloseError", err, err)
		}
	case <-time.After(2 * time.Second):
		t.Error("client took too long to disconnect")
	}
}

func TestNewServer(t *testing.T) {
	relay, err := nostr.RelayConnect(context.Background(), "ws://127.0.0.1:7447")
	if err != nil {
		panic(err)
	}
	sk := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(sk)
	nsec, _ := nip19.EncodePrivateKey(sk)
	npub, _ := nip19.EncodePublicKey(pub)
	t.Log("nsec: ", nsec)
	t.Log("npub: ", npub)
	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: time.Now(),
		Kind:      1,
		Tags:      nil,
		Content:   "sandy111 test golang relay",
	}

	// calling Sign sets the event ID field and the event Sig field
	ev.Sign(sk)
	status := relay.Publish(context.Background(), ev)
	t.Log(status)
}

func TestServer_Addr(t *testing.T) {
	nsec := "nsec1nhu4cvac38e4elxypma2nxajzgnrjnfpk5t27xz2urxxu7d0js0ss7whq7"
	npub := "npub1sh8576r3wudxhth24a2x59ezq7fu882akwg7zwa62rq7kt7dp5mqc7nh04"
	_, v, err := nip19.Decode(nsec)
	assert.NoError(t, err)
	t.Log(v.(string))

	_, v, err = nip19.Decode(npub)
	assert.NoError(t, err)
	t.Log(v.(string))

}
