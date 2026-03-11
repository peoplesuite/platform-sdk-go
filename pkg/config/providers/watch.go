package providers

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// WatchProvider loads config from a file and watches for changes.
type WatchProvider struct {
	Path     string
	Decoder  Decoder
	OnReload func() error

	watcher *fsnotify.Watcher
}

// NewWatch returns a WatchProvider that watches path and uses decoder.
func NewWatch(path string, decoder Decoder, reload func() error) (*WatchProvider, error) {

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &WatchProvider{
		Path:     path,
		Decoder:  decoder,
		OnReload: reload,
		watcher:  w,
	}, nil
}

// Name returns the provider name including the path.
func (p *WatchProvider) Name() string {
	return "watch:" + p.Path
}

// Load reads the file and decodes into cfg.
func (p *WatchProvider) Load(cfg any) error {

	data, err := readFile(p.Path)
	if err != nil {
		return err
	}

	if err := p.Decoder(data, cfg); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	return nil
}

// Start begins watching the file for changes.
func (p *WatchProvider) Start() error {

	dir := filepath.Dir(p.Path)

	if err := p.watcher.Add(dir); err != nil {
		return err
	}

	go p.loop()

	return nil
}

func (p *WatchProvider) loop() {

	for {
		select {

		case event, ok := <-p.watcher.Events:
			if !ok {
				return
			}

			if event.Name == p.Path && event.Op&(fsnotify.Write|fsnotify.Create) != 0 {

				log.Printf("config file changed: %s", event.Name)

				if p.OnReload != nil {
					if err := p.OnReload(); err != nil {
						log.Printf("config reload failed: %v", err)
					}
				}

			}

		case err, ok := <-p.watcher.Errors:
			if !ok {
				return
			}

			log.Printf("config watcher error: %v", err)
		}
	}
}
