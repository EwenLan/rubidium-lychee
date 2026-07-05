package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// msgHeaderRe matches lines like:
//   2026-07-05 10:00:00.123 [SEND] action round=5
//   2026-07-05 10:00:00.123 [RECV] start
var msgHeaderRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} \[(SEND|RECV)\] (\w+)(?: round=(\d+))?`)

// ParseLog reads a log file and extracts all message entries (two-line
// [DIR]/JSON records). Non-message log lines (INFO/DEBUG/etc.) are skipped.
func ParseLog(path string) ([]RecordedMessage, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log: %w", err)
	}
	defer f.Close()

	var out []RecordedMessage
	scanner := bufio.NewScanner(f)
	// Inquire messages can be large; allow up to 10 MB per line.
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if msgHeaderRe.FindStringSubmatch(line) == nil {
			continue
		}
		// Next line is the JSON envelope.
		if !scanner.Scan() {
			break
		}
		body := scanner.Bytes()
		if !json.Valid(body) {
			continue
		}
		var env struct {
			MsgName string `json:"msg_name"`
			MsgData struct {
				Round int `json:"round"`
			} `json:"msg_data"`
		}
		round := 0
		if json.Unmarshal(body, &env) == nil {
			round = env.MsgData.Round
		}
		out = append(out, RecordedMessage{
			Dir:      dirFromHeader(line),
			MsgName:  env.MsgName,
			Round:    round,
			Envelope: append([]byte(nil), body...),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan log: %w", err)
	}
	return out, nil
}

// dirFromHeader extracts the direction (SEND/RECV) from a header line.
func dirFromHeader(line string) string {
	m := msgHeaderRe.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	return m[1]
}

// FilterRecv returns only the RECV entries (server-sent messages from the
// client's perspective), suitable for replay by a mock server.
func FilterRecv(msgs []RecordedMessage) []RecordedMessage {
	var out []RecordedMessage
	for _, m := range msgs {
		if m.Dir == DirRecv {
			out = append(out, m)
		}
	}
	return out
}
