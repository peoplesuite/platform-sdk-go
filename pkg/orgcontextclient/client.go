package orgcontextclient

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	orgcontextv1 "peoplesuite/platform-contracts/gen/go/orgcontext/v1"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
)

// Config configures an Org Context client.
type Config struct {
	// Address is the gRPC address of the org-context service (host:port).
	Address string
	// Logger is optional; if nil no client-level logs are emitted.
	Logger *zap.Logger
	// Timeout is the default per-RPC timeout. If zero, callers should manage
	// deadlines via context when invoking RPCs.
	Timeout time.Duration
	// Pool is an optional shared client pool. If provided, connections are obtained
	// from the pool and the Client will not close them.
	Pool *sdkgrpc.ClientPool
}

// Client is a typed wrapper around orgcontextv1.OrgContextQueryServiceClient.
type Client struct {
	grpc   orgcontextv1.OrgContextQueryServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new Org Context client using the provided Config.
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
			Address:           cfg.Address,
			Logger:            cfg.Logger,
			Timeout:           cfg.Timeout,
			UnaryInterceptors: nil,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		grpc:   orgcontextv1.NewOrgContextQueryServiceClient(conn),
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
		grpc: orgcontextv1.NewOrgContextQueryServiceClient(conn),
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
func (c *Client) GRPC() orgcontextv1.OrgContextQueryServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}

// GetAssociateIDByEmail returns the associate ID for a person with the given email,
// or (0, false) if not found. It currently uses SearchAssociates and filters
// client-side by email address.
func (c *Client) GetAssociateIDByEmail(
	ctx context.Context,
	email string,
) (associateID int, found bool, err error) {
	if c == nil || c.grpc == nil {
		return 0, false, nil
	}

	resp, err := c.grpc.SearchAssociates(ctx, &orgcontextv1.SearchAssociatesRequest{})
	if err != nil {
		return 0, false, err
	}

	for _, assoc := range resp.GetAssociates() {
		person := assoc.GetPerson()
		if person == nil {
			continue
		}
		if strings.EqualFold(person.GetEmailAddress(), email) ||
			strings.EqualFold(person.GetBusinessEmailAddress(), email) {
			return int(assoc.GetAssociateId()), true, nil
		}
	}

	return 0, false, nil
}
