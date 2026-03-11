package runtime

import (
	"context"

	"go.uber.org/zap"
)

func (r *Runtime) startWorkers(ctx context.Context) {

	for _, w := range r.opts.Workers {

		go func(worker Worker) {

			if err := worker(ctx); err != nil {
				r.logger.Error("worker failed",
					zap.Error(err))
			}

		}(w)

	}
}
