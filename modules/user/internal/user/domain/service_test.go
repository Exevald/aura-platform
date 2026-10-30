package domain

import (
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	userevents "user/api/events"
)

type userRepositoryFake struct {
	users map[uuid.UUID]User
}

func (r *userRepositoryFake) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *userRepositoryFake) Store(user User) error {
	r.users[user.ID] = user
	return nil
}

func (r *userRepositoryFake) Find(id uuid.UUID) (User, error) {
	user, ok := r.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (r *userRepositoryFake) FindByEmail(email string) (User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (r *userRepositoryFake) Delete(id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

type userDispatcherFake struct {
	messages []proto.Message
}

func (d *userDispatcherFake) Dispatch(message proto.Message) error {
	d.messages = append(d.messages, message)
	return nil
}

func TestUserServiceLifecycle(t *testing.T) {
	repo := &userRepositoryFake{users: make(map[uuid.UUID]User)}
	dispatcher := &userDispatcherFake{}
	service := NewUserService(repo, dispatcher)

	id, err := service.CreateUser("Ada", "Ada", "Lovelace", "ada@example.com")
	require.NoError(t, err)
	require.IsType(t, &userevents.UserCreated{}, dispatcher.messages[0])

	user, err := service.GetUser(id)
	require.NoError(t, err)
	require.Equal(t, "Ada", user.DisplayName)

	err = service.UpdateUser(id, "Countess", "Ada", "Lovelace", "ada@example.com")
	require.NoError(t, err)
	require.Equal(t, "Countess", repo.users[id].DisplayName)
	require.IsType(t, &userevents.UserUpdated{}, dispatcher.messages[1])

	err = service.RemoveUser(id)
	require.NoError(t, err)
	require.IsType(t, &userevents.UserRemoved{}, dispatcher.messages[2])
	_, err = service.GetUser(id)
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserServiceRejectsInvalidData(t *testing.T) {
	service := NewUserService(
		&userRepositoryFake{users: make(map[uuid.UUID]User)},
		&userDispatcherFake{},
	)
	_, err := service.CreateUser("", "Ada", "Lovelace", "ada@example.com")
	require.ErrorIs(t, err, ErrInvalidUser)
	require.True(t, errors.Is(service.RemoveUser(uuid.Nil()), ErrInvalidUser))
}
