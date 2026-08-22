package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Event struct {
	TS        string `json:"ts"`
	Actor     string `json:"actor"`
	Tenant    string `json:"tenant"`
	Action    string `json:"action"`
	ObjectID  string `json:"object_id"`
	Outcome   string `json:"outcome"`
	RequestID string `json:"request_id"`
}

type Logger struct {
	mu sync.Mutex
}

func New() *Logger { return &Logger{} }

func (l *Logger) Emit(e Event) {
	if e.TS == "" {
		e.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	b = append(b, '\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = os.Stdout.Write(b)
}
