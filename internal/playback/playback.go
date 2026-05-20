package playback

import (
	"errors"
	"sync"
	"time"

	"github.com/MonsteRico/immich-frame/internal/cache"
)

type Asset struct {
	ID         string    `json:"id"`
	MediaURL   string    `json:"mediaUrl"`
	Title      string    `json:"title"`
	SourceName string    `json:"sourceName"`
	TakenAt    time.Time `json:"takenAt,omitempty"`
}

type State struct {
	Configured bool      `json:"configured"`
	Paused     bool      `json:"paused"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	Current    *Asset    `json:"current,omitempty"`
	Next       *Asset    `json:"next,omitempty"`
	Previous   *Asset    `json:"previous,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Queue struct {
	mu       sync.Mutex
	entries  []cache.Entry
	current  int
	previous []int
	paused   bool
	status   string
	message  string
}

func NewQueue(entries []cache.Entry) *Queue {
	q := &Queue{current: -1, status: "empty", message: "No photos are available yet."}
	q.Replace(entries)
	return q
}

func (q *Queue) Replace(entries []cache.Entry) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.entries = append([]cache.Entry(nil), entries...)
	q.previous = nil
	if len(q.entries) == 0 {
		q.current = -1
		q.status = "empty"
		q.message = "No photos are available yet."
		return
	}
	q.current = 0
	q.status = "ready"
	q.message = ""
}

func (q *Queue) Next() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.entries) == 0 {
		return "", errors.New("no cached media")
	}
	if q.current >= 0 {
		q.previous = append(q.previous, q.current)
		if len(q.previous) > 50 {
			q.previous = q.previous[1:]
		}
	}
	q.current = (q.current + 1) % len(q.entries)
	return q.entries[q.current].AssetID, nil
}

func (q *Queue) Previous() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.entries) == 0 {
		return "", errors.New("no cached media")
	}
	if len(q.previous) > 0 {
		last := len(q.previous) - 1
		q.current = q.previous[last]
		q.previous = q.previous[:last]
		return q.entries[q.current].AssetID, nil
	}
	q.current = (q.current - 1 + len(q.entries)) % len(q.entries)
	return q.entries[q.current].AssetID, nil
}

func (q *Queue) Pause() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.paused = true
}

func (q *Queue) Resume() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.paused = false
}

func (q *Queue) Paused() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.paused
}

func (q *Queue) State() State {
	q.mu.Lock()
	defer q.mu.Unlock()
	state := State{
		Configured: len(q.entries) > 0,
		Paused:     q.paused,
		Status:     q.status,
		Message:    q.message,
		UpdatedAt:  time.Now(),
	}
	if len(q.entries) == 0 || q.current < 0 {
		return state
	}
	current := asset(q.entries[q.current])
	state.Current = &current
	if len(q.entries) > 1 {
		next := asset(q.entries[(q.current+1)%len(q.entries)])
		state.Next = &next
	}
	if len(q.previous) > 0 {
		prev := asset(q.entries[q.previous[len(q.previous)-1]])
		state.Previous = &prev
	}
	return state
}

func asset(entry cache.Entry) Asset {
	return Asset{
		ID:         entry.AssetID,
		MediaURL:   "/media/" + entry.AssetID,
		Title:      entry.Title,
		SourceName: entry.SourceName,
		TakenAt:    entry.TakenAt,
	}
}
