package runtime

import "context"

// Module is a startable and stoppable component.
type Module interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// Modules is a list of Module started and stopped in order.
type Modules []Module

// Start starts all modules in order.
func (m Modules) Start(ctx context.Context) error {

	for _, mod := range m {
		if err := mod.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops all modules in reverse order.
func (m Modules) Stop(ctx context.Context) error {

	for _, mod := range m {
		if err := mod.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
