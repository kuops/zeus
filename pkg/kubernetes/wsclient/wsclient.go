package wsclient

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
)

type WSClient struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
}

func NewWSClient(conn *websocket.Conn) *WSClient {
	return &WSClient{conn: conn}
}

func (w *WSClient) Next() *remotecommand.TerminalSize {
	select {
	case size := <-w.sizeChan:
		return &size
	}
}

// Message
// Type      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (w *WSClient) Read(p []byte) (int, error) {
	var msg Message
	err := w.conn.ReadJSON(&msg)
	if err != nil {
		return 0, err
	}

	switch msg.Type {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		w.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown message type '%s'", msg.Type)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (w *WSClient) Write(p []byte) (n int, err error) {
	msg, err := json.Marshal(Message{
		Type: "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}
	w.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err = w.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}
