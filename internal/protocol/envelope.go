package protocol

import (
	"encoding/json"
	"fmt"

	"rubidium-lychee/internal/log"
	"rubidium-lychee/internal/transport"
)

// Envelope is the outer {msg_name, msg_data} wrapper used by every message.
type Envelope struct {
	MsgName string          `json:"msg_name"`
	MsgData json.RawMessage `json:"msg_data"`
}

// Message name constants.
const (
	MsgRegistration = "registration"
	MsgStart        = "start"
	MsgReady        = "ready"
	MsgInquire      = "inquire"
	MsgAction       = "action"
	MsgOver         = "over"
	MsgError        = "error"
)

// Send marshals v as msg_data and writes one framed envelope. The full
// envelope is logged to the log package (stdout + file if initialized).
func Send(c *transport.Conn, msgName string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", msgName, err)
	}
	env := Envelope{MsgName: msgName, MsgData: body}
	full, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope %s: %w", msgName, err)
	}
	log.LogMessage(log.DirSend, full)
	return c.WriteFrame(full)
}

// Recv reads one frame, decodes the envelope, and logs it. The raw frame
// bytes are logged (not the re-marshaled envelope) so the log reflects
// exactly what the server sent.
func Recv(c *transport.Conn) (Envelope, error) {
	body, err := c.ReadFrame()
	if err != nil {
		return Envelope{}, err
	}
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return Envelope{}, fmt.Errorf("unmarshal envelope: %w", err)
	}
	log.LogMessage(log.DirRecv, body)
	return env, nil
}

// SendRaw writes a pre-encoded envelope without re-marshaling. Used by the
// mock server in replay mode to forward recorded messages verbatim.
func SendRaw(c *transport.Conn, envelope []byte) error {
	log.LogMessage(log.DirSend, envelope)
	return c.WriteFrame(envelope)
}
