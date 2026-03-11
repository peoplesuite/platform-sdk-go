package settingsclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
	settingsv1 "peoplesuite/platform-contracts/gen/go/settings/v1"
)

// Config configures a Settings client.
type Config struct {
	// Address is the gRPC address of the settings service (host:port).
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

// Client is a typed wrapper around settingsv1.SettingsServiceClient.
type Client struct {
	grpc   settingsv1.SettingsServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new Settings client using the provided Config.
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
		grpc:   settingsv1.NewSettingsServiceClient(conn),
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
		grpc: settingsv1.NewSettingsServiceClient(conn),
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
func (c *Client) GRPC() settingsv1.SettingsServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}

// GetFeaturesForTenant returns feature definitions for the given tenant.
func (c *Client) GetFeaturesForTenant(
	ctx context.Context,
	tenantID string,
) ([]*settingsv1.FeatureDefinition, error) {
	if c == nil || c.grpc == nil {
		return nil, nil
	}

	resp, err := c.grpc.GetFeaturesForTenant(ctx, &settingsv1.GetFeaturesForTenantRequest{
		TenantId: tenantID,
	})
	if err != nil {
		return nil, err
	}

	return resp.Features, nil
}

// GetNavigation returns navigation items for the given tenant and context.
func (c *Client) GetNavigation(
	ctx context.Context,
	tenantID string,
	contextName string,
) ([]*settingsv1.NavItem, error) {
	if c == nil || c.grpc == nil {
		return nil, nil
	}

	if contextName == "" {
		contextName = "default"
	}

	resp, err := c.grpc.GetNavigation(ctx, &settingsv1.GetNavigationRequest{
		TenantId: tenantID,
		Context:  contextName,
	})
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// ListNotices returns notices for the given tenant.
func (c *Client) ListNotices(
	ctx context.Context,
	tenantID string,
) ([]*settingsv1.Notice, error) {
	if c == nil || c.grpc == nil {
		return nil, nil
	}

	resp, err := c.grpc.ListNotices(ctx, &settingsv1.ListNoticesRequest{
		TenantId: tenantID,
	})
	if err != nil {
		return nil, err
	}

	return resp.Notices, nil
}

// GetShellConfig returns whether admin shell is available for the tenant.
func (c *Client) GetShellConfig(
	ctx context.Context,
	tenantID string,
) (bool, error) {
	if c == nil || c.grpc == nil {
		return false, nil
	}

	resp, err := c.grpc.GetShellConfig(ctx, &settingsv1.GetShellConfigRequest{
		TenantId: tenantID,
	})
	if err != nil {
		return false, err
	}

	return resp.AdminAvailable, nil
}

// UpdateShellConfig updates admin shell availability for the tenant.
func (c *Client) UpdateShellConfig(
	ctx context.Context,
	tenantID string,
	adminAvailable bool,
) error {
	if c == nil || c.grpc == nil {
		return nil
	}

	_, err := c.grpc.UpdateShellConfig(ctx, &settingsv1.UpdateShellConfigRequest{
		TenantId:       tenantID,
		AdminAvailable: adminAvailable,
	})
	return err
}
