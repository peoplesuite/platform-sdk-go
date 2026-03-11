package profileclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
	profilev1 "peoplesuite/platform-contracts/gen/go/profile/v1"
)

// Config configures a Profile client.
type Config struct {
	Address string
	Logger  *zap.Logger
	Timeout time.Duration
	Pool    *sdkgrpc.ClientPool
}

// Client is a typed wrapper around profilev1.ProfileServiceClient.
type Client struct {
	grpc   profilev1.ProfileServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new Profile client using the provided Config.
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
		grpc:   profilev1.NewProfileServiceClient(conn),
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
		grpc: profilev1.NewProfileServiceClient(conn),
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
func (c *Client) GRPC() profilev1.ProfileServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}
