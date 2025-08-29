package confy

import (
	"sync"
	"sync/atomic"
)

// subID is an internal identifier for each subscriber
type subID uint64

var (
	mu      sync.RWMutex                // protects the subs map
	subs    = map[subID]chan struct{}{} // holds active subscribers (each with its own channel)
	counter atomic.Uint64               // generates unique subscriber IDs
)

// Subscribe registers a new listener and returns:
//   - a unique subscriber ID
//   - a receive-only channel that will get notified when changes happen
//
// The channel has a buffer size of 1 so that multiple notifications
// can be coalesced (no piling up if the receiver is slow).
func Subscribe() (subID, <-chan struct{}) {
	id := subID(counter.Add(1)) // atomically generate new ID
	ch := make(chan struct{}, 1)

	mu.Lock()
	subs[id] = ch
	mu.Unlock()

	return id, ch
}

// Unsubscribe removes a subscriber from the registry and closes its channel.
// Must be called when you no longer need the subscription, otherwise it will leak.
func Unsubscribe(id subID) {
	mu.Lock()
	if ch, ok := subs[id]; ok {
		delete(subs, id)
		close(ch) // safe: channel is unique to this subscriber
	}
	mu.Unlock()
}

// Notify broadcasts a signal to all active subscribers.
// It sends a struct{} into each channel in a non-blocking way:
//   - If the channel buffer is full, the send is skipped.
//   - This ensures Notify never blocks, even if a subscriber is slow or inactive.
func Notify() {
	mu.RLock()
	for _, ch := range subs {
		select {
		case ch <- struct{}{}:
			// successfully sent signal
		default:
			// skip if channel buffer already full
		}
	}
	mu.RUnlock()
}

// Close shuts down all subscribers at once.
// It closes every channel and clears the registry.
// After calling Close, no new notifications should be sent.
func Close() {
	mu.Lock()
	defer mu.Unlock()

	for id, ch := range subs {
		close(ch)
		delete(subs, id)
	}
}
