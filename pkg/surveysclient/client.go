package surveysclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	surveysv1 "peoplesuite/platform-contracts/gen/go/surveys/v1"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
)

// Config configures a Surveys client.
type Config struct {
	Address string
	Logger  *zap.Logger
	Timeout time.Duration
	Pool    *sdkgrpc.ClientPool
}

// Client is a typed wrapper around surveysv1.SurveysServiceClient.
type Client struct {
	grpc   surveysv1.SurveysServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new Surveys client using the provided Config.
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
		grpc:   surveysv1.NewSurveysServiceClient(conn),
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
		grpc: surveysv1.NewSurveysServiceClient(conn),
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
func (c *Client) GRPC() surveysv1.SurveysServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}
