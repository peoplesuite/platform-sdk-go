package frameworkclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	frameworkv1 "peoplesuite/platform-contracts/gen/go/framework/v1"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
)

// Config configures a Framework client.
type Config struct {
	Address string
	Logger  *zap.Logger
	Timeout time.Duration
	Pool    *sdkgrpc.ClientPool
}

// Client is a typed wrapper around frameworkv1.FrameworkServiceClient.
type Client struct {
	grpc   frameworkv1.FrameworkServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new Framework client using the provided Config.
func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, nil
	}

	var (
		conn *grpc.ClientConn
		err  error
	)

	if cfg.Pool != nil {
		conn, err = cfg.Pool.Get(cfg.Address)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = sdkgrpc.NewClientConn(sdkgrpc.ClientConfig{
			Address: cfg.Address,
			Logger:  cfg.Logger,
			Timeout: cfg.Timeout,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		grpc:   frameworkv1.NewFrameworkServiceClient(conn),
		conn:   conn,
		config: cfg,
	}, nil
}

// NewFromConn constructs a Client from an existing connection. The connection
// lifecycle is owned by the caller; Close is a no-op.
func NewFromConn(conn *grpc.ClientConn) *Client {
	if conn == nil {
		return nil
	}
	return &Client{
		grpc: frameworkv1.NewFrameworkServiceClient(conn),
		conn: nil,
	}
}

// Close closes the underlying connection if this client created it directly.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.config.Pool != nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GRPC returns the underlying generated gRPC client.
func (c *Client) GRPC() frameworkv1.FrameworkServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}
