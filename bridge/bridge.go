package bridge

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"time"

	"github.org/illicitonion/matrix-rssbot/event"
)

const prefix = "/_matrix/client/api/v1"

type bridge struct {
	Homeserver  string
	BotUser     string
	AccessToken string
	sendQueue   chan sendEventAttempt
}

func New(homeserver, accessToken, botUser string) *bridge {
	b := &bridge{
		Homeserver:  homeserver,
		BotUser:     botUser,
		AccessToken: accessToken,
		sendQueue:   make(chan sendEventAttempt, 10000),
	}

	go func() {
		for e := range b.sendQueue {
			if e.attempt > 0 {
				d := time.Duration(int64(math.Pow(1.5, float64(e.attempt)))) * time.Second
				time.Sleep(d)
			}
			err := b.actuallySend(e)
			if err != nil {
				log.Print(err)
			}
		}
	}()

	return b
}

func (b *bridge) SendMessage(roomID string, m *event.Message) {
	if !b.IsInRoom(roomID) {
		b.Join(roomID)
	}
	b.send(fmt.Sprintf("%s/rooms/%s/send/%s", prefix, roomID, "m.room.message"), m)
}

func (b *bridge) IsInRoom(roomID string) bool {
	return false
}

func (b *bridge) Join(roomID string) {
	b.send(fmt.Sprintf("%s/rooms/%s/join", prefix, roomID), struct{}{})
}

func (b *bridge) send(path string, content interface{}) {
	b.sendQueue <- sendEventAttempt{
		path: path,
		body: content,
	}
}

func (b *bridge) actuallySend(event sendEventAttempt) error {
	r, w := io.Pipe()
	enc := json.NewEncoder(w)
	go func() {
		if err := enc.Encode(event.body); err != nil {
			log.Printf("Error JSON-encoding transaction: %v", err)
		}
		w.Close()
	}()
	resp, err := http.Post(b.Homeserver+event.path+"?access_token="+b.AccessToken, "application/json", r)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		// Rate limited
		event.attempt += 1
		b.sendQueue <- event
		return fmt.Errorf("request to %v was rate limited", event.path)
	} else if resp.StatusCode != 200 {
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return fmt.Errorf("error dumping response after non-200 status code: %v", err)
		}
		return fmt.Errorf("non-200 from homeserver:\n%v", string(b))
	}
	return nil
}

type sendEventAttempt struct {
	path    string
	body    interface{}
	attempt int
}
