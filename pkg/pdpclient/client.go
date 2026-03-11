package pdpclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
	pdpv1 "peoplesuite/platform-contracts/gen/go/pdp/v1"
)

// Config configures a PDP client.
type Config struct {
	Address string
	Logger  *zap.Logger
	Timeout time.Duration
	Pool    *sdkgrpc.ClientPool
}

// Client is a typed wrapper around pdpv1.PdpServiceClient.
type Client struct {
	grpc   pdpv1.PdpServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new PDP client using the provided Config.
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
		grpc:   pdpv1.NewPdpServiceClient(conn),
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
		grpc: pdpv1.NewPdpServiceClient(conn),
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
func (c *Client) GRPC() pdpv1.PdpServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}
