package runtime

import (
	"context"
	"errors"
	"testing"
)

type fakeModule struct {
	startErr error
	stopErr  error
}

func (f *fakeModule) Start(ctx context.Context) error { return f.startErr }
func (f *fakeModule) Stop(ctx context.Context) error  { return f.stopErr }

func TestModules_Start_ReturnsFirstError(t *testing.T) {
	wantErr := errors.New("start failed")
	mods := Modules{
		&fakeModule{},
		&fakeModule{startErr: wantErr},
	}
	err := mods.Start(context.Background())
	if err != wantErr {
		t.Errorf("Start() = %v, want %v", err, wantErr)
	}
}

func TestModules_Start_Success(t *testing.T) {
	mods := Modules{&fakeModule{}, &fakeModule{}}
	if err := mods.Start(context.Background()); err != nil {
		t.Errorf("Start() = %v", err)
	}
}

func TestModules_Stop_ReturnsFirstError(t *testing.T) {
	wantErr := errors.New("stop failed")
	mods := Modules{
		&fakeModule{},
		&fakeModule{stopErr: wantErr},
	}
	err := mods.Stop(context.Background())
	if err != wantErr {
		t.Errorf("Stop() = %v, want %v", err, wantErr)
	}
}

func TestModules_Stop_Success(t *testing.T) {
	mods := Modules{&fakeModule{}, &fakeModule{}}
	if err := mods.Stop(context.Background()); err != nil {
		t.Errorf("Stop() = %v", err)
	}
}
