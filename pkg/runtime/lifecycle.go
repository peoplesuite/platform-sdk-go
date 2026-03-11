package runtime

import "context"

func runStartHooks(ctx context.Context, hooks []func(context.Context) error) error {

	for _, h := range hooks {
		if err := h(ctx); err != nil {
			return err
		}
	}

	return nil
}

func runStopHooks(ctx context.Context, hooks []func(context.Context) error) error {

	for _, h := range hooks {
		if err := h(ctx); err != nil {
			return err
		}
	}

	return nil
}
