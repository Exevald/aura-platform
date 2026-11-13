package domain

import (
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	authevents "auth/api/events"
)

type authClock struct {
	now time.Time
}

type mockCredentialRepository struct {
	credentials map[uuid.UUID]Credential
	findError   error
}

type mockPasswordHasher struct{}

type mockAuthDispatcher struct {
	messages []proto.Message
}

func TestAuthServiceCredentialLifecycle(t *testing.T) {
	repo := &mockCredentialRepository{credentials: make(map[uuid.UUID]Credential)}
	dispatcher := &mockAuthDispatcher{}
	service := newAuthServiceForTest(repo, dispatcher)
	userID := uuid.New()

	credential, err := service.CreateCredential(userID, "ada", "correct horse battery")
	require.NoError(t, err)
	require.Equal(t, "hash:correct horse battery", credential.PasswordHash)
	require.Empty(t, dispatcher.messages)

	_, err = service.CreateCredential(uuid.New(), "ada", "another valid password")
	require.ErrorIs(t, err, ErrCredentialExists)

	authenticated, err := service.Authenticate("ada", "correct horse battery")
	require.NoError(t, err)
	require.Equal(t, userID, authenticated.UserID)
	require.NoError(t, service.RecordLogin(authenticated))
	require.IsType(t, &authevents.UserLoggedIn{}, dispatcher.messages[0])

	err = service.ChangePassword(userID, "correct horse battery", "a completely new password")
	require.NoError(t, err)
	require.IsType(t, &authevents.PasswordChanged{}, dispatcher.messages[1])
	_, err = service.Authenticate("ada", "a completely new password")
	require.NoError(t, err)
}

func TestAuthServiceDoesNotHideRepositoryFailure(t *testing.T) {
	want := errors.New("database unavailable")
	repo := &mockCredentialRepository{credentials: make(map[uuid.UUID]Credential), findError: want}
	_, err := newAuthServiceForTest(repo, &mockAuthDispatcher{}).Authenticate(
		"ada",
		"correct horse battery",
	)
	require.ErrorIs(t, err, want)
}

func TestPasswordPolicy(t *testing.T) {
	require.ErrorIs(t, validatePassword("1234567"), ErrInvalidPassword)
	require.NoError(t, validatePassword("12345678"))
	require.NoError(t, validatePassword("very very long password"))
}

func (c authClock) Now() time.Time {
	return c.now
}

func (r *mockCredentialRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *mockCredentialRepository) Store(credential Credential) error {
	r.credentials[credential.ID] = credential
	return nil
}

func (r *mockCredentialRepository) Find(id uuid.UUID) (Credential, error) {
	credential, ok := r.credentials[id]
	if !ok {
		return Credential{}, ErrCredentialNotFound
	}
	return credential, nil
}

func (r *mockCredentialRepository) FindByLogin(login string) (Credential, error) {
	if r.findError != nil {
		return Credential{}, r.findError
	}
	for _, credential := range r.credentials {
		if credential.Login == login {
			return credential, nil
		}
	}
	return Credential{}, ErrCredentialNotFound
}

func (r *mockCredentialRepository) FindByUserID(userID uuid.UUID) (Credential, error) {
	for _, credential := range r.credentials {
		if credential.UserID == userID {
			return credential, nil
		}
	}
	return Credential{}, ErrCredentialNotFound
}

func (r *mockCredentialRepository) Delete(id uuid.UUID) error {
	delete(r.credentials, id)
	return nil
}

func (mockPasswordHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (mockPasswordHasher) Compare(hash, password string) error {
	if hash != "hash:"+password {
		return errors.New("mismatch")
	}
	return nil
}

func (d *mockAuthDispatcher) Dispatch(message proto.Message) error {
	d.messages = append(d.messages, message)
	return nil
}

func newAuthServiceForTest(repo *mockCredentialRepository, dispatcher *mockAuthDispatcher) AuthService {
	return NewAuthService(
		repo,
		mockPasswordHasher{},
		dispatcher,
		authClock{now: time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)},
	)
}
