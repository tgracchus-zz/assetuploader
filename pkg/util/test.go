package util

import (
	"context"
	"time"
)

// WaitUntilWithContext try to execute the action every waitTime for timeout time.
func WaitUntilWithContext(ctx context.Context, action func(ctx context.Context) error, waitTime time.Duration, timeout time.Duration) error {
	c := make(chan error, 1)
	defer close(c)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		go func() {
			err := action(ctx)
			select {
			case <-ctx.Done(): // timeout
				return
			default:
				c <- err // completed normally
				return
			}
		}()
		select {
		case <-ctx.Done():
			return ctx.Err() // timeout
		case err := <-c:
			if err == nil {
				return nil // completed normally
			}
		}
		<-time.After(waitTime) //waitTimeBeforeRetry
	}
}
