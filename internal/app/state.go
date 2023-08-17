package app

import (
	"sync"
	"time"
)

type state struct {
	mutex      sync.Mutex
	StartTime  time.Time `json:"start_time"`
	ErrorCount int64     `json:"error_count"`
}

func (s *state) IncrementErrorCount() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ErrorCount++
}
