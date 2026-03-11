package runtime

import (
	"context"
	"time"
)

var shutdownTimeout = 15 * time.Second

func (r *Runtime) shutdown(ctx context.Context) error {

	r.logger.Info("runtime shutting down")

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := runStopHooks(ctx, r.opts.StopHooks); err != nil {
		return err
	}

	return nil
}
