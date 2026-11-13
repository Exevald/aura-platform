package e2e

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accessapi "access/api/server/internalapi"
	authapi "auth/api/server/internalapi"
	sessionapi "session/api/server/internalapi"
	userapi "user/api/server/internalapi"
)

type fanOutGate struct {
	mu        sync.Mutex
	expected  int
	arrived   int
	armed     bool
	completed bool
	release   chan struct{}
}

type authRegistrationProxy struct {
	authapi.UnimplementedAuthServiceServer
	mu                sync.RWMutex
	client            authapi.AuthServiceClient
	failPrepareBefore atomic.Bool
	failPrepareAfter  atomic.Bool
	prepareGate       *fanOutGate
	commitGate        *fanOutGate
	rollbackGate      *fanOutGate
}

type userRegistrationProxy struct {
	userapi.UnimplementedUserServiceServer
	mu                  sync.RWMutex
	client              userapi.UserServiceClient
	failPrepareBefore   atomic.Bool
	failPrepareAfter    atomic.Bool
	failCommitRemaining atomic.Int32
	prepareGate         *fanOutGate
	commitGate          *fanOutGate
	rollbackGate        *fanOutGate
}

type sessionProxy struct {
	sessionapi.UnimplementedSessionServiceServer
	mu         sync.RWMutex
	client     sessionapi.SessionServiceClient
	failCreate atomic.Bool
}

type accessRegistrationProxy struct {
	accessapi.UnimplementedModuleAccessServiceServer
	mu                  sync.RWMutex
	client              accessapi.ModuleAccessServiceClient
	failPrepareBefore   atomic.Bool
	failPrepareAfter    atomic.Bool
	failCommitRemaining atomic.Int32
	prepareGate         *fanOutGate
	commitGate          *fanOutGate
	rollbackGate        *fanOutGate
}

func (p *authRegistrationProxy) PrepareCredentialRegistration(ctx context.Context, request *authapi.PrepareCredentialRegistrationRequest) (*authapi.PrepareCredentialRegistrationResponse, error) {
	if err := p.prepareGate.wait(ctx); err != nil {
		return nil, err
	}
	if p.failPrepareBefore.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected auth prepare failure")
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	response, err := client.PrepareCredentialRegistration(ctx, request)
	if err == nil && p.failPrepareAfter.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected lost auth prepare response")
	}
	return response, err
}

func (p *authRegistrationProxy) CommitCredentialRegistration(ctx context.Context, request *authapi.RegistrationTransactionRequest) (*authapi.RegistrationTransactionResponse, error) {
	if err := p.commitGate.wait(ctx); err != nil {
		return nil, err
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.CommitCredentialRegistration(ctx, request)
}

func (p *authRegistrationProxy) RollbackCredentialRegistration(ctx context.Context, request *authapi.RegistrationTransactionRequest) (*authapi.RegistrationTransactionResponse, error) {
	if err := p.rollbackGate.wait(ctx); err != nil {
		return nil, err
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.RollbackCredentialRegistration(ctx, request)
}

func (p *sessionProxy) CreateSession(ctx context.Context, request *sessionapi.CreateSessionRequest) (*sessionapi.CreateSessionResponse, error) {
	if p.failCreate.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected session failure")
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.CreateSession(ctx, request)
}

func (p *sessionProxy) ValidateSession(ctx context.Context, request *sessionapi.ValidateSessionRequest) (*sessionapi.ValidateSessionResponse, error) {
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.ValidateSession(ctx, request)
}

func (p *sessionProxy) RevokeSession(ctx context.Context, request *sessionapi.RevokeSessionRequest) (*sessionapi.RevokeSessionResponse, error) {
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.RevokeSession(ctx, request)
}

func (p *sessionProxy) RevokeUserSessions(ctx context.Context, request *sessionapi.RevokeUserSessionsRequest) (*sessionapi.RevokeUserSessionsResponse, error) {
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.RevokeUserSessions(ctx, request)
}

func (p *userRegistrationProxy) PrepareUserRegistration(ctx context.Context, request *userapi.PrepareUserRegistrationRequest) (*userapi.RegistrationTransactionResponse, error) {
	if err := p.prepareGate.wait(ctx); err != nil {
		return nil, err
	}
	if p.failPrepareBefore.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected user prepare failure")
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	response, err := client.PrepareUserRegistration(ctx, request)
	if err == nil && p.failPrepareAfter.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected lost user prepare response")
	}
	return response, err
}

func (p *userRegistrationProxy) CommitUserRegistration(ctx context.Context, request *userapi.RegistrationTransactionRequest) (*userapi.RegistrationTransactionResponse, error) {
	if err := p.commitGate.wait(ctx); err != nil {
		return nil, err
	}
	if remaining := p.failCommitRemaining.Load(); remaining > 0 && p.failCommitRemaining.CompareAndSwap(remaining, remaining-1) {
		return nil, status.Error(codes.Unavailable, "injected user commit failure")
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.CommitUserRegistration(ctx, request)
}

func (p *userRegistrationProxy) RollbackUserRegistration(ctx context.Context, request *userapi.RegistrationTransactionRequest) (*userapi.RegistrationTransactionResponse, error) {
	if err := p.rollbackGate.wait(ctx); err != nil {
		return nil, err
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.RollbackUserRegistration(ctx, request)
}

func (p *accessRegistrationProxy) PrepareModuleAccessRegistration(ctx context.Context, request *accessapi.PrepareModuleAccessRegistrationRequest) (*accessapi.RegistrationTransactionResponse, error) {
	if err := p.prepareGate.wait(ctx); err != nil {
		return nil, err
	}
	if p.failPrepareBefore.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected prepare failure")
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	response, err := client.PrepareModuleAccessRegistration(ctx, request)
	if err == nil && p.failPrepareAfter.Swap(false) {
		return nil, status.Error(codes.Unavailable, "injected lost prepare response")
	}
	return response, err
}

func (p *accessRegistrationProxy) CommitModuleAccessRegistration(ctx context.Context, request *accessapi.RegistrationTransactionRequest) (*accessapi.RegistrationTransactionResponse, error) {
	if err := p.commitGate.wait(ctx); err != nil {
		return nil, err
	}
	if remaining := p.failCommitRemaining.Load(); remaining > 0 && p.failCommitRemaining.CompareAndSwap(remaining, remaining-1) {
		return nil, status.Error(codes.Unavailable, "injected commit failure")
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.CommitModuleAccessRegistration(ctx, request)
}

func (p *accessRegistrationProxy) RollbackModuleAccessRegistration(ctx context.Context, request *accessapi.RegistrationTransactionRequest) (*accessapi.RegistrationTransactionResponse, error) {
	if err := p.rollbackGate.wait(ctx); err != nil {
		return nil, err
	}
	client, err := p.target()
	if err != nil {
		return nil, err
	}
	return client.RollbackModuleAccessRegistration(ctx, request)
}

func (g *fanOutGate) arm() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.armed {
		close(g.release)
	}
	g.arrived = 0
	g.armed = true
	g.completed = false
	g.release = make(chan struct{})
}

func (g *fanOutGate) reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.armed {
		close(g.release)
	}
	g.armed = false
}

func (g *fanOutGate) wasCompleted() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.completed
}

func (g *fanOutGate) wait(ctx context.Context) error {
	g.mu.Lock()
	if !g.armed {
		g.mu.Unlock()
		return nil
	}
	release := g.release
	g.arrived++
	if g.arrived == g.expected {
		g.completed = true
		g.armed = false
		close(release)
	}
	g.mu.Unlock()

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-release:
		return nil
	case <-ctx.Done():
		g.cancel(release)
		return ctx.Err()
	case <-timer.C:
		g.cancel(release)
		return status.Error(codes.DeadlineExceeded, "fan-out gate timed out")
	}
}

func (g *fanOutGate) cancel(release chan struct{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.armed && g.release == release {
		g.armed = false
		close(release)
	}
}

func (p *authRegistrationProxy) setClient(client authapi.AuthServiceClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = client
}

func (p *authRegistrationProxy) target() (authapi.AuthServiceClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.client == nil {
		return nil, status.Error(codes.Unavailable, "auth target is not ready")
	}
	return p.client, nil
}

func (p *sessionProxy) setClient(client sessionapi.SessionServiceClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = client
}

func (p *sessionProxy) target() (sessionapi.SessionServiceClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.client == nil {
		return nil, status.Error(codes.Unavailable, "session target is not ready")
	}
	return p.client, nil
}

func (p *userRegistrationProxy) setClient(client userapi.UserServiceClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = client
}

func (p *userRegistrationProxy) target() (userapi.UserServiceClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.client == nil {
		return nil, status.Error(codes.Unavailable, "user target is not ready")
	}
	return p.client, nil
}

func (p *accessRegistrationProxy) setClient(client accessapi.ModuleAccessServiceClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = client
}

func (p *accessRegistrationProxy) target() (accessapi.ModuleAccessServiceClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.client == nil {
		return nil, status.Error(codes.Unavailable, "access target is not ready")
	}
	return p.client, nil
}

func newFanOutGate(expected int) *fanOutGate {
	return &fanOutGate{expected: expected}
}
