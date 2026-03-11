package identityclient

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	identityv1 "peoplesuite/platform-contracts/gen/go/identity/v1"

	sdkgrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
)

// Config configures an Identity client.
type Config struct {
	// Address is the gRPC address of the identity service (host:port).
	Address string
	// Logger is optional; if nil no client-level logs are emitted.
	Logger *zap.Logger
	// Timeout is the default per-RPC timeout. If zero, a reasonable default should
	// be supplied by the caller when invoking RPCs.
	Timeout time.Duration
	// Pool is an optional shared client pool. If provided, connections are obtained
	// from the pool and the Client will not close them.
	Pool *sdkgrpc.ClientPool
}

// Client is a typed wrapper around identityv1.IdentityServiceClient.
//
// It exposes commonly used higher-level helpers while still allowing direct
// access to the underlying gRPC client.
type Client struct {
	grpc   identityv1.IdentityServiceClient
	conn   *grpc.ClientConn
	config Config
}

// New creates a new Identity client using the provided Config.
// If Config.Pool is set, the connection is obtained from the pool and the
// client will not close it. Otherwise it dials using sdkgrpc.NewClientConn.
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
		grpc:   identityv1.NewIdentityServiceClient(conn),
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
		grpc: identityv1.NewIdentityServiceClient(conn),
		conn: nil,
	}
}

// Close closes the underlying connection if this client created it directly.
// If the client was constructed from a pool or via NewFromConn, Close is a no-op.
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
func (c *Client) GRPC() identityv1.IdentityServiceClient {
	if c == nil {
		return nil
	}
	return c.grpc
}

// ResolveIdentity is a convenience helper for the ResolveIdentity RPC.
// It returns the resolved user ID and whether a new user was created.
func (c *Client) ResolveIdentity(
	ctx context.Context,
	providerHint, subject, email, username, givenName, familyName, source string,
) (userID string, created bool, err error) {
	if c == nil || c.grpc == nil {
		return "", false, nil
	}

	provider := issuerToIdentityProvider(providerHint)
	req := &identityv1.ResolveIdentityRequest{
		Provider:        provider,
		ProviderSubject: subject,
		Email:           email,
		Username:        username,
		GivenName:       givenName,
		FamilyName:      familyName,
		Source:          source,
	}

	resp, err := c.grpc.ResolveIdentity(ctx, req)
	if err != nil {
		return "", false, err
	}

	if resp.User != nil {
		userID = resp.User.UserId
	}

	return userID, resp.Created, nil
}

// GetUser is a convenience helper for the GetUser RPC.
func (c *Client) GetUser(ctx context.Context, userID string) (*identityv1.User, error) {
	if c == nil || c.grpc == nil {
		return nil, nil
	}

	resp, err := c.grpc.GetUser(ctx, &identityv1.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return resp.User, nil
}

// GetAssociateIDForUser is a convenience helper for the GetAssociateIDForUser RPC.
// It returns (associateID, found, error).
func (c *Client) GetAssociateIDForUser(
	ctx context.Context,
	userID string,
) (int32, bool, error) {
	if c == nil || c.grpc == nil {
		return 0, false, nil
	}

	resp, err := c.grpc.GetAssociateIDForUser(ctx, &identityv1.GetAssociateIDForUserRequest{
		UserId: userID,
	})
	if err != nil {
		return 0, false, err
	}

	return resp.AssociateId, resp.Found, nil
}

// issuerToIdentityProvider maps an issuer hint to an IdentityProvider enum.
func issuerToIdentityProvider(issuerHint string) identityv1.IdentityProvider {
	s := strings.ToLower(strings.TrimSpace(issuerHint))

	switch {
	case strings.Contains(s, "keycloak"), strings.Contains(s, "authentik"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_KEYCLOAK
	case strings.Contains(s, "google"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_GOOGLE
	case strings.Contains(s, "microsoft"), strings.Contains(s, "azure"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_MICROSOFT
	case strings.Contains(s, "github"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB
	case strings.Contains(s, "linkedin"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_LINKEDIN
	case strings.Contains(s, "telegram"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_TELEGRAM
	case strings.Contains(s, "slack"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_SLACK
	case strings.Contains(s, "teams"):
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_TEAMS
	default:
		return identityv1.IdentityProvider_IDENTITY_PROVIDER_KEYCLOAK
	}
}
