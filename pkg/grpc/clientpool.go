package grpc

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ClientPool caches gRPC client connections by address.
type ClientPool struct {
	mu      sync.RWMutex
	clients map[string]*grpc.ClientConn
	logger  *zap.Logger
	cfg     ClientConfig
}

// NewClientPool returns a new ClientPool with the given config.
func NewClientPool(cfg ClientConfig) *ClientPool {
	return &ClientPool{
		clients: make(map[string]*grpc.ClientConn),
		logger:  cfg.Logger,
		cfg:     cfg,
	}
}

// Get returns a gRPC connection for the given address.
// If it does not exist it will create one.
func (p *ClientPool) Get(address string) (*grpc.ClientConn, error) {

	p.mu.RLock()

	conn, ok := p.clients[address]

	p.mu.RUnlock()

	if ok {
		return conn, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// double check after lock
	if conn, ok := p.clients[address]; ok {
		return conn, nil
	}

	cfg := p.cfg
	cfg.Address = address

	newConn, err := NewClientConn(cfg)
	if err != nil {
		return nil, fmt.Errorf("grpc client connect: %w", err)
	}

	p.clients[address] = newConn

	if p.logger != nil {
		p.logger.Info("gRPC client pooled",
			zap.String("address", address),
		)
	}

	return newConn, nil
}

// Close closes all connections in the pool.
func (p *ClientPool) Close() error {

	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error

	for addr, conn := range p.clients {

		if err := conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}

		if p.logger != nil {
			p.logger.Info("gRPC client closed",
				zap.String("address", addr),
			)
		}
	}

	p.clients = map[string]*grpc.ClientConn{}

	return firstErr
}
