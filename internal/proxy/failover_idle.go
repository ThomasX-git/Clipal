package proxy

import (
	"context"
	"errors"
	"time"
)

var errUpstreamIdleTimeout = errors.New("upstream idle timeout")

func isUpstreamIdleTimeout(ctx context.Context, err error) bool {
	if err == nil {
		// Some callers only have the attempt context; in that case the idle timeout is
		// conveyed via context cancellation cause.
		return errors.Is(context.Cause(ctx), errUpstreamIdleTimeout)
	}
	if errors.Is(err, errUpstreamIdleTimeout) {
		return true
	}
	if errors.Is(err, context.Canceled) && errors.Is(context.Cause(ctx), errUpstreamIdleTimeout) {
		return true
	}
	return false
}

func stopTimer(t *time.Timer) {
	if t == nil {
		return
	}
	t.Stop()
}
