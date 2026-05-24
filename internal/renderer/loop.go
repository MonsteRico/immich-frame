package renderer

import (
	"context"
	"errors"
	"time"
)

type FetchSnapshot func(context.Context) (Snapshot, error)
type DecodeAsset func(context.Context, Asset) error

type Loop struct {
	Fetch  FetchSnapshot
	Decode DecodeAsset
	Now    func() time.Time

	visible *Asset
	status  string
	message string
	err     error
}

type LoopState struct {
	Visible   *Asset
	Status    string
	Message   string
	Err       error
	UpdatedAt time.Time
}

func (l *Loop) Step(ctx context.Context) LoopState {
	now := time.Now
	if l.Now != nil {
		now = l.Now
	}
	if l.Fetch == nil {
		return l.fail(errors.New("renderer fetch function is not configured"), now())
	}
	snapshot, err := l.Fetch(ctx)
	if err != nil {
		return l.fail(err, now())
	}
	l.status = snapshot.Status
	l.message = snapshot.Message
	l.err = nil
	if snapshot.Current == nil {
		return l.state(now())
	}
	if l.visible != nil && l.visible.ID == snapshot.Current.ID {
		return l.state(now())
	}
	if l.Decode != nil {
		if err := l.Decode(ctx, *snapshot.Current); err != nil {
			return l.fail(err, now())
		}
	}
	current := *snapshot.Current
	l.visible = &current
	return l.state(now())
}

func (l *Loop) fail(err error, updatedAt time.Time) LoopState {
	l.err = err
	if l.status == "" {
		l.status = "degraded"
	}
	if l.message == "" {
		l.message = err.Error()
	}
	return l.state(updatedAt)
}

func (l *Loop) state(updatedAt time.Time) LoopState {
	var visible *Asset
	if l.visible != nil {
		copy := *l.visible
		visible = &copy
	}
	return LoopState{
		Visible:   visible,
		Status:    l.status,
		Message:   l.message,
		Err:       l.err,
		UpdatedAt: updatedAt,
	}
}
