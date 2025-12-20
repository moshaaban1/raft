package replication

import (
	"context"
	"time"
)

func startTicker(ctx context.Context, interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
			// TODO: call fn()

			case <-ctx.Done():

				return
			}
		}
	}()
}
