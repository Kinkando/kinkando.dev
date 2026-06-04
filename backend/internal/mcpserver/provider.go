package mcpserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// UserResolver maps a Firebase UID to the internal app user UUID.
type UserResolver interface {
	GetIDByFirebaseUID(ctx context.Context, firebaseUID string) (uuid.UUID, error)
}

// Provider builds and caches a per-user MCP server and in-process client
// session. It satisfies the gemini.SessionProvider interface.
type Provider struct {
	base     BaseDeps
	resolver UserResolver

	mu       sync.Mutex
	servers  map[string]*mcp.Server
	sessions map[string]*mcp.ClientSession
}

// NewProvider creates a Provider backed by the given base dependencies and user
// resolver. Call Close when the server shuts down to release all sessions.
func NewProvider(base BaseDeps, resolver UserResolver) *Provider {
	return &Provider{
		base:     base,
		resolver: resolver,
		servers:  make(map[string]*mcp.Server),
		sessions: make(map[string]*mcp.ClientSession),
	}
}

// ServerFor returns the cached MCP server for firebaseUID, creating one on the
// first call. Returns an error if the user cannot be resolved.
func (p *Provider) ServerFor(ctx context.Context, firebaseUID string) (*mcp.Server, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if s, ok := p.servers[firebaseUID]; ok {
		return s, nil
	}

	userUUID, err := p.resolver.GetIDByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("mcpserver provider: resolve user %q: %w", firebaseUID, err)
	}
	if userUUID == (uuid.UUID{}) {
		return nil, fmt.Errorf("mcpserver provider: user %q not found; sign in via the web app first", firebaseUID)
	}

	srv := New(Deps{
		BaseDeps:    p.base,
		UserUUID:    userUUID,
		FirebaseUID: firebaseUID,
	})
	p.servers[firebaseUID] = srv
	return srv, nil
}

// SessionFor returns the cached in-process MCP client session for firebaseUID,
// creating one (and the backing server) on the first call.
func (p *Provider) SessionFor(ctx context.Context, firebaseUID string) (*mcp.ClientSession, error) {
	// Fast path without the lock.
	p.mu.Lock()
	if sess, ok := p.sessions[firebaseUID]; ok {
		p.mu.Unlock()
		return sess, nil
	}
	p.mu.Unlock()

	// ServerFor acquires the lock internally.
	srv, err := p.ServerFor(ctx, firebaseUID)
	if err != nil {
		return nil, err
	}

	serverT, clientT := mcp.NewInMemoryTransports()

	serverSession, err := srv.Connect(ctx, serverT, nil)
	if err != nil {
		return nil, fmt.Errorf("mcpserver provider: connect server session for %q: %w", firebaseUID, err)
	}

	cli := mcp.NewClient(&mcp.Implementation{Name: "kinkando-in-process", Version: "0.1.0"}, nil)
	clientSession, err := cli.Connect(ctx, clientT, nil)
	if err != nil {
		serverSession.Close() //nolint:errcheck
		return nil, fmt.Errorf("mcpserver provider: connect client session for %q: %w", firebaseUID, err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Another goroutine may have raced us.
	if existing, ok := p.sessions[firebaseUID]; ok {
		// We built a duplicate; close ours and return the winner.
		clientSession.Close() //nolint:errcheck
		serverSession.Close() //nolint:errcheck
		return existing, nil
	}

	p.sessions[firebaseUID] = clientSession
	return clientSession, nil
}

// Close releases all cached in-process client sessions.
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for uid, sess := range p.sessions {
		if err := sess.Close(); err != nil {
			p.base.Logger.Sugar().Warnf("mcpserver provider: close session for %q: %v", uid, err)
		}
	}
	p.sessions = make(map[string]*mcp.ClientSession)
	return nil
}
