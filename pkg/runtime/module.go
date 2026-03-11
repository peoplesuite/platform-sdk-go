package runtime

import "context"

type Module interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type Modules []Module

func (m Modules) Start(ctx context.Context) error {

	for _, mod := range m {
		if err := mod.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (m Modules) Stop(ctx context.Context) error {

	for _, mod := range m {
		if err := mod.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
