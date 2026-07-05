package log

import (
	"encoding/json"
	"fmt"
)

// Direction constants for message logging.
const (
	DirSend = "SEND" // client → server
	DirRecv = "RECV" // server → client
)

// LogMessage writes a message envelope to the log in a two-line format:
//
//	TIMESTAMP [DIR] msg_name round=N
//	{compact JSON envelope}
//
// The round is extracted from msg_data.round if present. If the logger is
// not initialized, this is a no-op.
func LogMessage(dir string, envelope []byte) {
	var env struct {
		MsgName string `json:"msg_name"`
		MsgData struct {
			Round int `json:"round"`
		} `json:"msg_data"`
	}
	round := 0
	if json.Unmarshal(envelope, &env) == nil {
		round = env.MsgData.Round
	}
	roundTag := ""
	if round > 0 {
		roundTag = fmt.Sprintf(" round=%d", round)
	}
	write(fmt.Sprintf("%s [%s] %s%s\n", timestamp(), dir, env.MsgName, roundTag))
	write(string(envelope) + "\n")
}

// RecordedMessage is a parsed message entry from a log file.
type RecordedMessage struct {
	Dir      string          // "SEND" or "RECV"
	MsgName  string          // from envelope
	Round    int             // from msg_data.round (0 if absent)
	Envelope json.RawMessage // raw envelope JSON
}
