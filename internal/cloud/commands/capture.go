package commands

import "sync"

// CaptureQueue holds user-requested snapshot captures per device.
type CaptureQueue struct {
	mu      sync.Mutex
	pending map[string]bool
}

// NewCaptureQueue creates an empty capture command queue.
func NewCaptureQueue() *CaptureQueue {
	return &CaptureQueue{pending: make(map[string]bool)}
}

// Request queues a one-shot capture for the device.
func (q *CaptureQueue) Request(deviceID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pending[deviceID] = true
}

// Take returns true once if a capture was requested, then clears the flag.
func (q *CaptureQueue) Take(deviceID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.pending[deviceID] {
		return false
	}
	delete(q.pending, deviceID)
	return true
}

// Pending reports whether a capture is waiting for the next heartbeat.
func (q *CaptureQueue) Pending(deviceID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pending[deviceID]
}
